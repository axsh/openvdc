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

var LxcConfigFile string
var ContainerName string

type Settings struct {
	ScriptPath      string
	LinuxUpScript   string
	LinuxDownScript string
	BridgeName      string
	OvsUpScript     string
	OvsDownScript   string
	OvsName         string
	TapName         string
}

type NetworkInterface struct {
	Type     string
	Ipv4Addr string
	MacAddr  string
	TapName  string
}

var settings Settings

var interfaces []NetworkInterface

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

func (p *LXCHypervisorProvider) CreateDriver(containerName string) (hypervisor.HypervisorDriver, error) {
	return &LXCHypervisorDriver{
		log:     log.WithFields(log.Fields{"hypervisor":"lxc", "instance_id": containerName}),
		lxcpath: lxc.DefaultConfigPath(),
		name:    containerName,
		// Set pre-defined template option from gopkg.in/lxc/go-lxc.v2/options.go
		template: lxc.DownloadTemplateOptions,
	}, nil
}

type LXCHypervisorDriver struct {
	log       *log.Entry
	imageName string
	hostName  string
	lxcpath   string
	template  lxc.TemplateOptions
	name      string
}

func (d *LXCHypervisorDriver) modifyConf() error {
	f, err := ioutil.ReadFile(LxcConfigFile)

	if err != nil {
		return errors.Wrapf(err, "Failed loading %s", LxcConfigFile)
	}

	cf := cleanConfigFile(string(f))

	var newSettings string
	//for i, _ := range interfaces {
	newSettings = updateSettings(interfaces[0], newSettings)
	//}

	result := strings.Join([]string{cf, newSettings}, "")
	err = ioutil.WriteFile(LxcConfigFile, []byte(result), 0644)
	if err != nil {
		return errors.Wrapf(err, "Couldn't save %s", LxcConfigFile)
	}
	d.log.Debug(result)
	return nil
}

func updateSettings(nwi NetworkInterface, input string) string {

	settings.TapName = nwi.TapName

	output := input + "\n"

	output += fmt.Sprintf("lxc.network.veth.pair=%s\n", nwi.TapName)

	if nwi.Ipv4Addr != "" {
		output += fmt.Sprintf("lxc.network.ipv4=%s\n", nwi.Ipv4Addr)
	}

	if nwi.MacAddr != "" {
		output += fmt.Sprintf("lxc.network.hwaddr=%s\n", nwi.MacAddr)
	}

	containerPath := filepath.Join(lxc.DefaultConfigPath(), ContainerName)

	output += fmt.Sprintf("lxc.network.script.up=%s\n", filepath.Join(containerPath, "up.sh"))
	output += fmt.Sprintf("lxc.network.script.down=%s\n", filepath.Join(containerPath, "down.sh"))

	switch nwi.Type {
	case "linux":
		bridgeConnect := fmt.Sprintf("brctl addif %s %s \n", settings.BridgeName, settings.TapName)

		generateScriptFromTemplate(settings.LinuxUpScript, "up.sh", bridgeConnect)
		generateScriptFromTemplate(settings.LinuxDownScript, "down.sh", bridgeConnect)

	case "ovs":
		bridgeConnect := fmt.Sprintf("ovs-vsctl add-port %s %s \n", settings.OvsName, settings.TapName)

		generateScriptFromTemplate(settings.OvsUpScript, "up.sh", bridgeConnect)
		generateScriptFromTemplate(settings.OvsDownScript, "down.sh", bridgeConnect)

	default:
		log.Fatalf("Unrecognized bridge type.")
	}

	return output
}

func generateScriptFromTemplate(scriptTemplate string, generatedScriptName string, bridgeConnect string) {
	f, err := ioutil.ReadFile(filepath.Join(settings.ScriptPath, scriptTemplate))

	if err != nil {
		log.Warnln("Failed loading script template: ", err)
	}

	var output string = "#!/bin/sh\n"

	if f != nil {
		if generatedScriptName == "up.sh" {
			output = bridgeConnect + string(f)
		}
	} else {

		output = bridgeConnect
	}

	containerPath := filepath.Join(lxc.DefaultConfigPath(), ContainerName)

	err = ioutil.WriteFile(filepath.Join(containerPath, generatedScriptName), []byte(output), 0755)

	if err != nil {
		log.Fatalln("Failed saving generated script to container path: ", err)
	}
}

func cleanConfigFile(input string) string {
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		if strings.Contains(line, "lxc.network.link") {
			lines[i] = ""
		}
	}

	output := strings.Join(lines, "\n")

	return output
}

func resetConfigFile() {
	f, err := ioutil.ReadFile(LxcConfigFile)

	if err != nil {
		log.Fatalf("Failed loading lxc default.conf: ", err)
	}

	lines := strings.Split(string(f), "\n")

	for i, line := range lines {
		if strings.Contains(line, "hwaddr") ||
			strings.Contains(line, "ipv4") ||
			strings.Contains(line, "script.up") ||
			strings.Contains(line, "script.down") ||
			strings.Contains(line, "veth.pair") {
			lines[i] = ""
		}
	}

	output := strings.Join(lines, "\n")

	err = ioutil.WriteFile(LxcConfigFile, []byte(output), 0644)
	if err != nil {
		log.Fatalln("Couldn't restore lxc config file", err)
	}
}

func (d *LXCHypervisorDriver) CreateInstance(i *model.Instance, in model.ResourceTemplate) error {

	lxcTmpl, ok := in.(*model.LxcTemplate)

	if !ok {

		d.log.Fatalf("BUGON: Unsupported model type: %T", in)

	}

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	LxcConfigFile = c.ConfigFileName()

	ContainerName = d.name

	if err != nil {
		return errors.Wrap(err, "Failed lxc.NewContainer")
	}

	d.log.Infoln("Creating lxc-container...")

	if err := c.Create(d.template); err != nil {
		return errors.Wrap(err, "Failed lxc.Create")
	}

	loadConfigFile()

	for _, i := range lxcTmpl.GetInterfaces() {
		interfaces = append(interfaces,
			NetworkInterface{
				Type:     i.GetBridge(),
				Ipv4Addr: i.GetIpv4Addr(),
				MacAddr:  i.GetMacaddr(),
				TapName:  d.name,
			},
		)
	}

	if err := d.modifyConf(); err != nil {
		return err
	}

	interfaces = nil

	return nil

}

func loadConfigFile() {
	settings.ScriptPath = viper.GetString("hypervisor.script-path")
	settings.LinuxUpScript = viper.GetString("bridges.linux.up-script")
	settings.LinuxDownScript = viper.GetString("bridges.linux.down-script")
	settings.BridgeName = viper.GetString("bridges.linux.name")
	settings.OvsUpScript = viper.GetString("bridges.ovs.up-script")
	settings.OvsDownScript = viper.GetString("bridges.ovs.down-script")
	settings.OvsName = viper.GetString("bridges.ovs.name")
}

func (d *LXCHypervisorDriver) DestroyInstance() error {
	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		return errors.Wrap(err, "Failed lxc.NewContainer")
	}

	if c.State() == lxc.RUNNING {
		d.log.Infoln("Stopping lxc-container..")
		if err := c.Stop(); err != nil {
			return errors.Wrap(err, "Failed lxc.Stop")
		}
	}

	d.log.Infoln("Destroying lxc-container..")
	if err := c.Destroy(); err != nil {
		return errors.Wrap(err, "Failed lxc.Destroy")
	}
	return nil
}

func (d *LXCHypervisorDriver) StartInstance() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		return errors.Wrap(err, "Failed lxc.NewContainer")
	}

	d.log.Infoln("Starting lxc-container...")
	if err := c.Start(); err != nil {
		return errors.Wrap(err, "Failed lxc.Start")
	}

	d.log.Infoln("Waiting for lxc-container to become RUNNING")
	if ok := c.Wait(lxc.RUNNING, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for RUNNING")
	}
	return nil
}

func (d *LXCHypervisorDriver) StopInstance() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		return errors.Wrap(err, "Failed lxc.NewContainer")
	}

	d.log.Infoln("Stopping lxc-container..")
	if err := c.Stop(); err != nil {
		return errors.Wrap(err, "Failed lxc.Stop")
	}

	d.log.Infoln("Waiting for lxc-container to become STOPPED")
	if ok := c.Wait(lxc.STOPPED, 30*time.Second); !ok {
		return errors.New("Failed or timedout to wait for STOPPED")
	}
	return nil
}

func (d *LXCHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &lxcConsole{
		lxc: d,
	}
}

func (d *LXCHypervisorDriver) newContainer() (*lxc.Container, error) {
	return lxc.NewContainer(d.name, d.lxcpath)
}

type lxcConsole struct {
	lxc *LXCHypervisorDriver
	attached *os.Process
	wg *sync.WaitGroup
	pty *os.File
	tty string
}

func (con *lxcConsole) Attach(stdin io.Reader, stdout, stderr io.Writer) error {
	c, err := con.lxc.newContainer()
	if err != nil {
		return errors.Wrap(err, "lxc.NewContainer")
	}

	if c.State() != lxc.RUNNING {
		return errors.New("lxc-container can not perform console")
	}

  return con.attachShell(c, stdin, stdout, stderr)
  //return con.console(c, stdin, stdout, stderr)
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

func (con *lxcConsole) attachShell(c *lxc.Container, stdin io.Reader, stdout, stderr io.Writer) error {
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

	pid, err := c.RunCommandNoWait([]string{"/bin/bash"}, options)
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

func (con *lxcConsole) console(c *lxc.Container, stdin io.Reader, stdout, stderr io.Writer) error {
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

	if err := c.Console(options); err != nil {
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
