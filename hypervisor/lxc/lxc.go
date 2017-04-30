// +build linux

package lxc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"text/template"
	"time"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/kr/pty"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

type BridgeType int

const (
	None BridgeType = iota // 0
	Linux
	OVS
)

func (t BridgeType) String() string {
	switch t {
	case Linux:
		return "linux"
	case OVS:
		return "ovs"
	default:
		return "none"
	}
}

var settings struct {
	ScriptPath      string
	BridgeName      string
	BridgeType      BridgeType
	LinuxUpScript   string
	LinuxDownScript string
	OvsUpScript     string
	OvsDownScript   string
}

func init() {
	hypervisor.RegisterProvider("lxc", &LXCHypervisorProvider{})
	viper.SetDefault("hypervisor.script-path", "/etc/openvdc/scripts")
	// Default script file names in pkg/conf/scripts/*
	viper.SetDefault("bridges.linux.up-script", "linux-bridge-up.sh.tmpl")
	viper.SetDefault("bridges.linux.down-script", "linux-bridge-down.sh.tmpl")
	viper.SetDefault("bridges.ovs.up-script", "ovs-up.sh.tmpl")
	viper.SetDefault("bridges.ovs.down-script", "ovs-down.sh.tmpl")
}

type LXCHypervisorProvider struct {
}

func (p *LXCHypervisorProvider) Name() string {
	return "lxc"
}

func (p *LXCHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	if sub.IsSet("bridges.name") {
		settings.BridgeName = sub.GetString("bridges.name")
		if sub.IsSet("bridges.type") {
			switch sub.GetString("bridges.type") {
			case "linux":
				settings.BridgeType = Linux
			case "ovs":
				settings.BridgeType = OVS
			default:
				return errors.Errorf("Unknown bridges.type value: %s", sub.GetString("bridges.type"))
			}
		}
	} else if sub.IsSet("bridges.linux.name") {
		log.Warn("bridges.linux.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.linux.name")
		settings.BridgeType = Linux
	} else if sub.IsSet("bridges.ovs.name") {
		log.Warn("bridges.ovs.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.ovs.name")
		settings.BridgeType = OVS
	}

	// They have default value.
	settings.ScriptPath = sub.GetString("hypervisor.script-path")
	settings.LinuxUpScript = sub.GetString("bridges.linux.up-script")
	settings.LinuxDownScript = sub.GetString("bridges.linux.down-script")
	settings.OvsUpScript = sub.GetString("bridges.ovs.up-script")
	settings.OvsDownScript = sub.GetString("bridges.ovs.down-script")
	return nil
}

func (p *LXCHypervisorProvider) CreateDriver(instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	lxcTmpl, ok := template.(*model.LxcTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.LxcTemplate: %T", template)
	}
	containerName := instance.GetId()
	c, err := lxc.NewContainer(containerName, lxc.DefaultConfigPath())
	if err != nil {
		return nil, errors.Wrap(err, "Failed lxc.NewContainer")
	}

	driver := &LXCHypervisorDriver{
		Base: hypervisor.Base{
			Log:      log.WithFields(log.Fields{"hypervisor": "lxc", "instance_id": containerName}),
			Instance: instance,
		},
		template:  lxcTmpl,
		container: c,
	}
	return driver, nil
}

type LXCHypervisorDriver struct {
	hypervisor.Base
	template  *model.LxcTemplate
	imageName string
	hostName  string
	container *lxc.Container
}

func (h *LXCHypervisorDriver) log() *log.Entry {
	return h.Base.Log
}

const lxcNetworkTemplate = `
lxc.network.type=veth
lxc.network.flags=up
lxc.network.veth.pair={{.TapName}}
lxc.network.script.up={{.UpScript}}
lxc.network.script.down={{.DownScript}}
{{- with .IFace.Ipv4Addr}}
lxc.network.ipv4={{$.IFace.Ipv4Addr}}
{{- end}}
{{- with .IFace.Macaddr}}
lxc.network.hwaddr={{$.IFace.Macaddr}}
{{- end}}
`

func (d *LXCHypervisorDriver) modifyConf() error {
	lxcconf, err := os.OpenFile(d.container.ConfigFileName(), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return errors.Wrapf(err, "Failed opening %s", d.container.ConfigFileName())
	}
	defer lxcconf.Close()

	// Start to append lxc.network entries to tmp file.

	// Append comment header
	fmt.Fprintf(lxcconf, "\n# OpenVDC Network Configuration\n")

	if lxc.VersionAtLeast(1, 1, 0) {
		/*
			https://linuxcontainers.org/lxc/manpages/man5/lxc.container.conf.5.html
			lxc.network
			may be used without a value to clear all previous network options.

			However requires the change 6b0d5538.
			https://github.com/lxc/lxc/commit/6b0d553864a16462850d87d4d2e9056ea136ebad
		*/
		fmt.Fprintf(lxcconf, "# Here clear all network options.\nlxc.network=\n")
	} else {
		/*
			lxc.network.type with no value does same thing.
			https://github.com/lxc/lxc/blob/stable-1.0/src/lxc/confile.c#L369-L377
		*/
		fmt.Fprintf(lxcconf, "# Here clear all network options.\nlxc.network.type=\n")
	}
	nwTemplate, err := template.New("lxc.network").Parse(lxcNetworkTemplate)
	if err != nil {
		errors.Wrap(err, "Failed to parse lxc.network template")
	}

	d.log().Debug("resource:", resource)

	if len(resource.Interfaces) > 0 && settings.BridgeType == None {
		d.log().Errorf("Network interfaces are requested to create but no bridge is configured")
	} else {
		// Write lxc.network.* entries.
		for idx, i := range d.template.Interfaces {
			tval := struct {
				IFace      *model.LxcTemplate_Interface
				TapName    string
				UpScript   string
				DownScript string
				IFIndex    int
			}{
				IFace:      i,
				IFIndex:    idx,
				TapName:    fmt.Sprintf("%s_%02d", d.container.Name(), idx),
				UpScript:   filepath.Join(d.containerDir(), "up.sh"),
				DownScript: filepath.Join(d.containerDir(), "down.sh"),
			}
			if err := nwTemplate.Execute(lxcconf, tval); err != nil {
				return errors.Wrapf(err, "Failed to render lxc.network template: %v", tval)
			}
		}
		lxcconf.Sync()
	}

	if d.log().Level <= log.DebugLevel {
		buf, _ := ioutil.ReadFile(d.container.ConfigFileName())
		d.log().Debug(string(buf))
	}
	return nil
}

func (d *LXCHypervisorDriver) containerDir() string {
	containerDir, _ := filepath.Split(d.container.ConfigFileName())
	return containerDir
}

func (d *LXCHypervisorDriver) renderUpDownScript(scriptTemplate, generateScript string) error {
	tmplPath := filepath.Join(settings.ScriptPath, scriptTemplate)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse script template: %s", tmplPath)
	}
	genPath := filepath.Join(d.containerDir(), generateScript)
	gen, err := os.OpenFile(genPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to create up/down script: %s", genPath)
	}
	defer gen.Close()
	err = tmpl.Execute(gen, settings)
	if err != nil {
		return errors.Wrapf(err, "Failed to render up/down script: %s", genPath)
	}
	return nil
}

var lxcArch = map[string]string{
	"amd64": "amd64",
	"386":   "i386",
	// TODO: powerpc, arm, arm64
}

func (d *LXCHypervisorDriver) CreateInstance() error {
	d.log().Infoln("Creating lxc-container...")
	lxcTmpl := d.template.GetLxcTemplate()
	template := lxc.TemplateOptions{
		Template:  lxcTmpl.Template,
		Arch:      lxcTmpl.Arch,
		ExtraArgs: lxcTmpl.ExtraArgs,
	}
	switch lxcTmpl.Template {
	case "download":
		template.Distro = lxcTmpl.Distro
		template.Release = lxcTmpl.Release
		template.Variant = lxcTmpl.Variant
	default:
		template.Release = lxcTmpl.Release
	}

	if template.Arch == "" {
		// Guess LXC Arch name
		template.Arch = lxcArch[runtime.GOARCH]
		if template.Arch == "" {
			return errors.Errorf("Unable to guess LXC arch name")
		}
	}

	if err := d.container.Create(template); err != nil {
		return errors.Wrap(err, "Failed lxc.Create")
	}

	if err := d.modifyConf(); err != nil {
		return err
	}

	// Force reload the modified container config.
	d.container.ClearConfig()
	if err := d.container.LoadConfigFile(d.container.ConfigFileName()); err != nil {
		return errors.Wrap(err, "Failed lxc.LoadConfigFile")
	}

	switch settings.BridgeType {
	case None:
		// Do nothing
	case Linux:
		if err := d.renderUpDownScript(settings.LinuxUpScript, "up.sh"); err != nil {
			return err
		}
		if err := d.renderUpDownScript(settings.LinuxDownScript, "down.sh"); err != nil {
			return err
		}
	case OVS:
		if err := d.renderUpDownScript(settings.OvsUpScript, "up.sh"); err != nil {
			return err
		}
		if err := d.renderUpDownScript(settings.OvsDownScript, "down.sh"); err != nil {
			return err
		}
	default:
		log.Fatalf("BUGON: Unknown bridge type: %s", settings.BridgeType)
	}
	return nil

}

func (d *LXCHypervisorDriver) DestroyInstance() error {
	if d.container.State() == lxc.RUNNING {
		d.log().Infoln("Stopping lxc-container..")
		if err := d.container.Stop(); err != nil {
			return errors.Wrap(err, "Failed lxc.Stop")
		}
	}

	d.log().Infoln("Destroying lxc-container..")
	if err := d.container.Destroy(); err != nil {
		return errors.Wrap(err, "Failed lxc.Destroy")
	}
	return nil
}

func (d *LXCHypervisorDriver) StartInstance() error {
	d.log().Infoln("Starting lxc-container...")
	if err := d.container.Start(); err != nil {
		return errors.Wrap(err, "Failed lxc.Start")
	}

	d.log().Infoln("Waiting for lxc-container to become RUNNING")
	if ok := d.container.Wait(lxc.RUNNING, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for RUNNING")
	}
	return nil
}

func (d *LXCHypervisorDriver) StopInstance() error {
	d.log().Infoln("Stopping lxc-container..")
	if err := d.container.Stop(); err != nil {
		return errors.Wrap(err, "Failed lxc.Stop")
	}

	d.log().Infoln("Waiting for lxc-container to become STOPPED")
	if ok := d.container.Wait(lxc.STOPPED, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for STOPPED")
	}
	return nil
}

func (d *LXCHypervisorDriver) RebootInstance() error {
	d.log().Infoln("Rebooting lxc-container..")
	if err := d.container.Reboot(); err != nil {
		return errors.Wrap(err, "Failed lxc.Reboot")
	}
	return nil
}

func (d *LXCHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &lxcConsole{
		lxc: d,
	}
}

type lxcConsole struct {
	lxc      *LXCHypervisorDriver
	attached *os.Process
	fds      []*os.File
	tty      string
}

func (con *lxcConsole) container() *lxc.Container {
	return con.lxc.container
}

func (con *lxcConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args)
}

func (con *lxcConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, []string{"/bin/bash"})
}

func (con *lxcConsole) pipeAttach(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	if con.container().State() != lxc.RUNNING {
		return nil, errors.New("lxc-container can not perform console")
	}

	fds := make([]*os.File, 6)
	closeAll := func() {
		for _, fd := range fds {
			fd.Close()
		}
	}

	var err error
	rIn, wIn, err := os.Pipe() // stdin
	if err != nil {
		return nil, errors.Wrap(err, "Failed os.Pipe for stdin")
	}
	fds = append(fds, rIn, wIn)
	con.fds = append(con.fds, wIn)
	rOut, wOut, err := os.Pipe() // stdout
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stdout")
	}
	fds = append(fds, rOut, wOut)
	con.fds = append(con.fds, rOut)
	rErr, wErr, err := os.Pipe() // stderr
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stderr")
	}
	fds = append(fds, rErr, wErr)
	con.fds = append(con.fds, rErr)

	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed, 3)
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(wIn, param.Stdin)
		if err == nil {
			// param.Stdin was closed due to EOF so needs to send EOF to pipe as well
			wIn.Close()
		}
		// TODO: handle err is not nil case
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stdout, rOut)
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stderr, rErr)
		closeChan <- err
		defer waitClosed.Done()
	}()
	go func() {
		waitClosed.Wait()
		defer close(closeChan)
	}()

	err = con.attachShell(rIn, wOut, wErr, param.Envs, args)
	if err != nil {
		defer func() {
			closeAll()
			con.fds = []*os.File{}
		}()
		return nil, err
	}

	// Close file descriptors for child process.
	defer func() {
		rIn.Close()
		wOut.Close()
		wErr.Close()
	}()

	return closeChan, nil
}

func (con *lxcConsole) AttachPty(param *hypervisor.ConsoleParam, ptyreq *hypervisor.SSHPtyReq) (<-chan hypervisor.Closed, error) {
	if con.container().State() != lxc.RUNNING {
		return nil, errors.New("lxc-container can not perform console")
	}

	fpty, ftty, err := pty.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open tty")
	}
	// Close primary socket
	defer ftty.Close()
	con.fds = append(con.fds, fpty)

	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed, 3)
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(fpty, param.Stdin)
		if err == nil {
			// param.Stdin was closed due to EOF so needs to send EOF to pty as well
			fpty.Close()
		}
		// TODO: handle err is not nil case
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stdout, fpty)
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stderr, fpty)
		closeChan <- err
		defer waitClosed.Done()
	}()
	go func() {
		waitClosed.Wait()
		defer close(closeChan)
	}()

	SetWinsize(ftty.Fd(), &Winsize{Height: uint16(ptyreq.Rows), Width: uint16(ptyreq.Columns)})
	/*
		modes, err := TcGetAttr(ftty.Fd())
		if err != nil {
			return errors.Wrap(err, "Failed TcGetAttr")
		}
		modes.Lflag &^= syscall.ECHO
		modes.Iflag |= syscall.IGNCR
		err = TcSetAttr(ftty.Fd(), modes)
		if err != nil {
			return errors.Wrap(err, "Failed TcSetAttr")
		}
	*/

	if ptyreq.Term != "" {
		param.Envs["TERM"] = ptyreq.Term
	}
	err = con.attachShell(ftty, ftty, ftty, param.Envs, []string{"/bin/bash"})
	if err != nil {
		defer func() {
			fpty.Close()
			con.fds = []*os.File{}
		}()
		return nil, err
	}
	con.tty = ftty.Name()
	return closeChan, nil
	//return con.console(stdin, stdout, stderr)
}

func (con *lxcConsole) UpdateWindowSize(w, h uint32) error {
	if !(len(con.fds) == 1) {
		return errors.New("tty is not opened")
	}
	return SetWinsize(con.fds[0].Fd(), &Winsize{
		Width:  uint16(w),
		Height: uint16(h),
	})
}

type consoleWaitError struct {
	os.ProcessState
}

func (c *consoleWaitError) Error() string {
	return c.ProcessState.String()
}

func (c *consoleWaitError) ExitCode() int {
	// http://stackoverflow.com/questions/10385551/get-exit-code-go
	if status, ok := c.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus()
	}
	log.Warnf("This platform %s does not support syscall.WaitStatus", runtime.GOOS)
	if !c.Success() {
		return 1
	}
	return 0
}

func (con *lxcConsole) Wait() error {
	if con.attached == nil {
		return errors.New("No process is found")
	}
	defer func() {
		log := log.WithField("pid", con.attached.Pid)
		if con.tty != "" {
			log = log.WithField("tty", con.tty)
		}

		for _, fd := range con.fds {
			fd.Close()
		}
		con.fds = nil
		con.attached = nil
		con.tty = ""
		log.Info("Closed attached session")
	}()

	state, err := con.attached.Wait()
	if err != nil {
		con.attached.Release()
		return errors.Wrap(err, "Failed Process.Wait")
	}

	if !state.Success() {
		return &consoleWaitError{
			ProcessState: *state,
		}
	}
	return nil
}

func (con *lxcConsole) ForceClose() error {
	if con.attached == nil {
		return errors.New("No process is found")
	}
	// This sends just signal. pty and pid are
	// closed by Wait()
	return errors.WithStack(con.attached.Kill())
}

func (con *lxcConsole) attachShell(stdin, stdout, stderr *os.File, envs map[string]string, args []string) error {
	options := lxc.DefaultAttachOptions
	options.StdinFd = stdin.Fd()
	options.StdoutFd = stdout.Fd()
	options.StderrFd = stderr.Fd()
	options.ClearEnv = true
	if len(envs) > 0 {
		s := make([]string, len(envs))
		for k, v := range envs {
			s = append(s, fmt.Sprintf("%s=%s", k, v))
		}
		options.Env = s
	}

	pid, err := con.container().RunCommandNoWait(args, options)
	if err != nil {
		return errors.Wrap(err, "Failed lxc.RunCommandNoWait")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "Failed os.FindProcess: %d", pid)
	}
	con.attached = proc

	return nil
}

func (con *lxcConsole) console(stdin, stdout, stderr *os.File) error {
	options := lxc.DefaultConsoleOptions
	options.StdinFd = stdin.Fd()
	options.StdoutFd = stdout.Fd()
	options.StderrFd = stderr.Fd()
	options.EscapeCharacter = '~'

	if err := con.container().Console(options); err != nil {
		return errors.Wrap(err, "Failed lxc.Console")
	}
	return nil
}

// https://github.com/creack/termios/blob/master/raw/raw.go
func TcSetAttr(fd uintptr, termios *syscall.Termios) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}

// https://github.com/creack/termios/blob/master/raw/raw.go
func TcGetAttr(fd uintptr) (*syscall.Termios, error) {
	var termios = &syscall.Termios{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios))); err != 0 {
		return nil, err
	}
	return termios, nil
}

//https://github.com/creack/termios/blob/master/win/win.go
// Winsize stores the Heighty and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

func SetWinsize(fd uintptr, ws *Winsize) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws))); err != 0 {
		return err
	}
	return nil
}
