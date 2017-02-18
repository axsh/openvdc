// +build linux

package lxc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"
	"unsafe"

	log "github.com/Sirupsen/logrus"
  "github.com/kr/pty"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
	lxc "gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type BridgeType int
const (
	None BridgeType = iota // 0
	Linux
	OVS
)

func (t BridgeType) String() string {
	switch t {
		Linux:
		return "linux"
		OVS:
		return "ovs"
		default:
		return "none"
	}
}

var settings struct {
	ScriptPath      string
	BridgeName      string
	BridgeType 			BridgeType
	LinuxUpScript   string
	LinuxDownScript string
	OvsUpScript     string
	OvsDownScript   string
}

func init() {
	hypervisor.RegisterProvider("lxc", &LXCHypervisorProvider{})
	viper.SetDefault("hypervisor.script-path", "/etc/openvdc/scripts")
	// Default script file names in pkg/conf/scripts/*
	viper.SetDefault("bridges.linux.up-script", "linux-bridge-up.sh")
	viper.SetDefault("bridges.linux.down-script", "linux-bridge-down.sh")
	viper.SetDefault("bridges.ovs.up-script", "ovs-up.sh")
	viper.SetDefault("bridges.ovs.down-script", "ovs-down.sh")
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
				case "linux"
				settings.BridgeType = Linux
				case "ovs"
				settings.BridgeType = OVS
				default:
				return errors.Errorf("Unknown bridges.type value: %s", sub.GetString("bridges.type"))
			}
		}
	}else	if sub.IsSet("bridges.linux.name") {
		log.Warn("bridges.linux.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.linux.name")
		settings.BridgeType = Linux
	}else	if sub.IsSet("bridges.ovs.name") {
		log.Warn("bridges.ovs.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.ovs.name")
		settings.BridgeType = OVS
	}

	if settings.BridgeName == "" {
		return errors.New("Missing bridges.name parameter")
	}

	if settings.BridgeType == None {
		return errors.New("Missing bridges.type parameter")
	}

	// They have default value.
	settings.ScriptPath = sub.GetString("hypervisor.script-path")
	settings.LinuxUpScript = sub.GetString("bridges.linux.up-script")
	settings.LinuxDownScript = sub.GetString("bridges.linux.down-script")
	settings.OvsUpScript = sub.GetString("bridges.ovs.up-script")
	settings.OvsDownScript = sub.GetString("bridges.ovs.down-script")
}

func (p *LXCHypervisorProvider) CreateDriver(containerName string) (hypervisor.HypervisorDriver, error) {
	c, err := lxc.NewContainer(containerName, lxc.DefaultConfigPath())
	if err != nil {
		return nil, errors.Wrap(err, "Failed lxc.NewContainer")
	}

	return &LXCHypervisorDriver{
		log:     log.WithFields(log.Fields{"hypervisor":"lxc", "instance_id": containerName}),
		// Set pre-defined template option from gopkg.in/lxc/go-lxc.v2/options.go
		template: lxc.DownloadTemplateOptions,
		container: c,
	}, nil
}

type LXCHypervisorDriver struct {
	log       *log.Entry
	imageName string
	hostName  string
	template  lxc.TemplateOptions
	container	*lxc.Container
}


const lxcNetworkTemplate := `
lxc.network.type=veth
lxc.network.veth.pair={{.TapName}}
lxc.network.script.up={{.UpScript}}
lxc.network.script.down={{.DownScript}}
{{with .IFace.Ipv4Addr}}
lxc.network.ipv4={{.IFace.Ipv4Addr}}
{{- end}}
{{with .IFace.MacAddr}}
lxc.network.hwaddr={{.IFace.MacAddr}}
{{- end}}
`

func (d *LXCHypervisorDriver) modifyConf(resource *model.LxcTemplate) error {
	lxcconf, err := os.OpenFile(d.container.ConfigPath(), os.O_WRONLY | os.O_APPEND, 0)
	if err != nil {
		return errors.Wrapf(err, "Failed opening %s", d.container.ConfigPath())
	}
	defer lxcconf.Close()

	// Append lxc.network entries to tmp file.

	/* Append comment header and lxc.network with no parameter
			https://linuxcontainers.org/lxc/manpages/man5/lxc.container.conf.5.html
			lxc.network
			may be used without a value to clear all previous network options.
	*/
	fmt.Fprintf(lxcconf, "\n# OpenVDC Network Configuration\n\n# Here clear all network options.\nlxc.network=\n")
	nwTemplate, err := template.New("lxc.network").Parse(lxcNetworkTemplate)
	if err != nil {
		errors.Wrap(err, "Failed to parse lxc.network template")
	}

	// Write lxc.network.* entries.
	for idx, i := range resource.Interfaces {
		var tval := struct {
			IFace *model.LxcTemplate_Interface,
			TapName string,
			UpScript string,
			DownScript string,
			IFIndex int,
		}{
			IFace: i,
			IFIndex: idx,
			TapName: d.container.Name(),
		}
		if err := nwTemplate.Execute(lxcconf, tval); err != nil {
			return errors.Wrapf(err, "Failed to render lxc.network template: %v", tval)
		}
	}
	lxcconf.Sync()

	if d.log.Level <= log.DebugLevel {
		buf, _ := ioutil.ReadFile(d.container.ConfigPath())
		d.log.Debug(string(buf))
	}
	return nil
}

func (d *LXCHypervisorDriver) renderUpDownScript(scriptTemplate, generateScript string) error {
	containerDir, _ := filepath.Split(d.container.ConfigPath())
	tmplPath := filepath.Join(settings.ScriptPath, scriptTemplate)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse script template: %s", tmplPath)
	}
	genPath := filepath.Join(containerDir, generateScript)
	gen, err := os.OpenFile(genPath, os.O_CREATE | os.O_WRONLY, 0755)
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

func (d *LXCHypervisorDriver) CreateInstance(i *model.Instance, in model.ResourceTemplate) error {

	lxcTmpl, ok := in.(*model.LxcTemplate)

	if !ok {

		d.log.Fatalf("BUGON: Unsupported model type: %T", in)

	}

	d.log.Infoln("Creating lxc-container...")
	if err := d.container.Create(d.template); err != nil {
		return errors.Wrap(err, "Failed lxc.Create")
	}

	if err := d.modifyConf(lxcTmpl); err != nil {
		return err
	}

	switch settings.BridgeType {
		case Linux
			if err := d.renderUpDownScript(settings.LinuxUpScript, "up.sh"); err != nil {
				return err
			}
			if err := d.renderUpDownScript(settings.LinuxDownScript, "down.sh"); err != nil {
				return err
			}
		case OVS
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
		d.log.Infoln("Stopping lxc-container..")
		if err := d.container.Stop(); err != nil {
			return errors.Wrap(err, "Failed lxc.Stop")
		}
	}

	d.log.Infoln("Destroying lxc-container..")
	if err := d.container.Destroy(); err != nil {
		return errors.Wrap(err, "Failed lxc.Destroy")
	}
	return nil
}

func (d *LXCHypervisorDriver) StartInstance() error {
	d.log.Infoln("Starting lxc-container...")
	if err := d.container.Start(); err != nil {
		return errors.Wrap(err, "Failed lxc.Start")
	}

	d.log.Infoln("Waiting for lxc-container to become RUNNING")
	if ok := d.container.Wait(lxc.RUNNING, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for RUNNING")
	}
	return nil
}

func (d *LXCHypervisorDriver) StopInstance() error {
	d.log.Infoln("Stopping lxc-container..")
	if err := d.container.Stop(); err != nil {
		return errors.Wrap(err, "Failed lxc.Stop")
	}

	d.log.Infoln("Waiting for lxc-container to become STOPPED")
	if ok := d.container.Wait(lxc.STOPPED, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for STOPPED")
	}
	return nil
}

func (d *LXCHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &lxcConsole{
		lxc: d,
	}
}

type lxcConsole struct {
	lxc *LXCHypervisorDriver
	attached *os.Process
	wg *sync.WaitGroup
	pty *os.File
	tty string
}

func (con *lxcConsole) container() *lxc.Container {
	return con.lxc.container
}

func (con *lxcConsole) Attach(stdin io.Reader, stdout, stderr io.Writer) error {
	if con.container().State() != lxc.RUNNING {
		return errors.New("lxc-container can not perform console")
	}

  return con.attachShell(stdin, stdout, stderr)
  //return con.console(stdin, stdout, stderr)
}

func (con *lxcConsole) Wait() error {
	if con.attached == nil {
		return errors.New("No process is found")
	}
	defer func() {
		err := con.pty.Close()
		log.WithFields(log.Fields{
			"tty": con.tty,
			"pid": con.attached.Pid,
		}).WithError(errors.WithStack(err)).Info("TTY session closed")
		con.pty = nil
		con.attached = nil
	}()

	_, err := con.attached.Wait()
	if err != nil {
		con.attached.Release()
	}
	return errors.Wrap(err, "Failed Process.Wait")
}

func (con *lxcConsole) ForceClose() error {
	if con.attached == nil {
		return errors.New("No process is found")
	}
	// This sends just signal. pty and pid are
	// closed by Wait()
	return errors.WithStack(con.attached.Kill())
}

func (con *lxcConsole) attachShell(stdin io.Reader, stdout, stderr io.Writer) error {
	var wg sync.WaitGroup
	fpty, ftty, err := pty.Open()
	if err != nil {
		return errors.Wrapf(err, "Failed to open tty")
	}
	// Close primary socket
	defer ftty.Close()

	wg.Add(1)
	go func(){
		defer wg.Done()
		io.Copy(fpty, stdin)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(stdout, fpty)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(stderr, fpty)
	}()

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

	options := lxc.DefaultAttachOptions
	options.StdinFd	= ftty.Fd()
	options.StdoutFd = ftty.Fd()
	options.StderrFd = ftty.Fd()
	options.ClearEnv = true

	pid, err := con.container().RunCommandNoWait([]string{"/bin/bash"}, options)
	if err != nil {
		err = errors.WithStack(err)
		con.lxc.log.WithError(err).Errorln("Failed to AttachShell")
		defer fpty.Close()
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		err = errors.WithStack(err)
		con.lxc.log.WithError(err).Errorf("Failed to find attached shell process: %d", pid)
		defer fpty.Close()
		return err
	}
	con.attached = proc
	con.wg = &wg
	con.pty = fpty
	con.tty = ftty.Name()
	return nil
}

func (con *lxcConsole) console(stdin io.Reader, stdout, stderr io.Writer) error {
	fpty, ftty, err := pty.Open()
	if err != nil {
		return errors.Wrapf(err, "Failed to open tty")
	}
	defer ftty.Close()
	defer fpty.Close()

	go io.Copy(fpty, stdin)
	go io.Copy(stdout, fpty)
	go io.Copy(stderr, os.Stderr)

	options := lxc.DefaultConsoleOptions
	options.StdinFd					= ftty.Fd()
	options.StdoutFd				= ftty.Fd()
	options.StderrFd				= ftty.Fd()
	options.EscapeCharacter = '~'

	if err := con.container().Console(options); err != nil {
		err = errors.WithStack(err)
		con.lxc.log.WithError(err).Error("Failed lxc.Console")
		return err
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
