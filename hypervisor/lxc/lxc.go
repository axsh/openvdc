// +build linux

package lxc

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/kr/pty"
	"github.com/pkg/errors"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

func init() {
	hypervisor.RegisterProvider("lxc", &LXCHypervisorProvider{})
}

type LXCHypervisorProvider struct {
}

func (p *LXCHypervisorProvider) Name() string {
	return "lxc"
}

func (d *LXCHypervisorDriver) GetContainerState(i *model.Instance) (hypervisor.ContainerState, error) {

	c, err := lxc.NewContainer(d.name, d.lxcpath)

	if err != nil {
		return hypervisor.ContainerState_NONE, err
	}

	var containerState hypervisor.ContainerState

	switch c.State() {
	case lxc.STOPPED:
		containerState = hypervisor.ContainerState_STOPPED
	case lxc.STARTING:
		containerState = hypervisor.ContainerState_STARTING
	case lxc.RUNNING:
		containerState = hypervisor.ContainerState_RUNNING
	case lxc.STOPPING:
		containerState = hypervisor.ContainerState_STOPPING
	case lxc.ABORTING:
		containerState = hypervisor.ContainerState_ABORTING
	default:
		containerState = hypervisor.ContainerState_NONE
	}

	return containerState, err
}

func (p *LXCHypervisorProvider) CreateDriver(containerName string) (hypervisor.HypervisorDriver, error) {
	return &LXCHypervisorDriver{
		log:     log.WithField("hypervisor", "lxc"),
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

func (d *LXCHypervisorDriver) CreateInstance(i *model.Instance, in model.ResourceTemplate) error {

	lxcTmpl, ok := in.(*model.LxcTemplate)

	if !ok {

		log.Fatal("BUGON: Unsupported model type")

	}

	c, err := lxc.NewContainer(d.name, d.lxcpath)

	if err != nil {

		d.log.Errorln(err)

		return err

	}

	d.log.Infoln("Creating lxc-container...")

	if err := c.Create(d.template); err != nil {

		d.log.Errorln(err)

		return err

	}

	var conf string

	for _, i := range lxcTmpl.GetInterfaces() {

		if i.GetIpv4Addr() == "" {

			conf += fmt.Sprintf("lxc.network.ipv4=%s\n", i.GetIpv4Addr())

		}

		if i.GetMacaddr() == "" {

			conf += fmt.Sprintf("lxc.network.hwaddr=%s\n", i.GetMacaddr())

		}

	}

	path := c.ConfigFileName()

	f, err := os.OpenFile(path, os.O_WRONLY, 0)

	defer f.Close()

	_, err = f.WriteString(conf)

	return nil

}

func (d *LXCHypervisorDriver) DestroyInstance() error {
	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		d.log.Errorln(err)
		return err
	}

	if c.State() == lxc.RUNNING {
		c.Stop()
	}

	d.log.Infoln("Destroying lxc-container..")
	if err := c.Destroy(); err != nil {
		d.log.Errorln(err)
		return err
	}
	return nil
}

func (d *LXCHypervisorDriver) StartInstance() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		d.log.Errorln(err)
		return err
	}

	d.log.Infoln("Starting lxc-container...")
	if err := c.Start(); err != nil {
		d.log.Errorln(err)
		return err
	}

	d.log.Infoln("Waiting for lxc-container to become RUNNING")
	if ok := c.Wait(lxc.RUNNING, 30*time.Second); !ok {
		d.log.Errorln("Failed or timedout to wait for RUNNING")
		return fmt.Errorf("Failed or timedout to wait for RUNNING")
	}
	return nil
}

func (d *LXCHypervisorDriver) StopInstance() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		d.log.Errorln(err)
		return err
	}

	d.log.Infoln("Stopping lxc-container..")
	if err := c.Stop(); err != nil {
		d.log.Errorln(err)
		return err
	}

	d.log.Infoln("Waiting for lxc-container to become STOPPED")
	if ok := c.Wait(lxc.STOPPED, 30*time.Second); !ok {
		d.log.Errorln("Failed or timedout to wait for STOPPED")
		return fmt.Errorf("Failed or timedout to wait for STOPPED")
	}
	return nil
}

func (d *LXCHypervisorDriver) RebootInstance() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		d.log.Errorln(err)
		return err
	}

	d.log.Infoln("Rebooting lxc-container..")
	if err := c.Reboot(); err != nil {
		d.log.Errorln(err)
		return err
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
	lxc      *LXCHypervisorDriver
	attached *os.Process
	wg       *sync.WaitGroup
	pty      *os.File
	tty      string
}

func (con *lxcConsole) Attach(stdin io.Reader, stdout, stderr io.Writer) error {
	c, err := con.lxc.newContainer()
	if err != nil {
		con.lxc.log.Errorln(err)
		return err
	}

	if c.State() != lxc.RUNNING {
		con.lxc.log.Errorf("lxc-container can not perform console")
		return fmt.Errorf("lxc-container can not perform console")
	}

	return con.attachShell(c, stdin, stdout, stderr)
	//return con.console(c, stdin, stdout, stderr)
}

func (con *lxcConsole) Wait() error {
	if con.attached == nil {
		return fmt.Errorf("No process is found")
	}
	defer func() {
		err := con.pty.Close()
		log.WithFields(log.Fields{
			"tty": con.tty,
			"pid": con.attached.Pid,
		}).WithError(err).Info("TTY session closed")
		con.pty = nil
		con.attached = nil
	}()

	_, err := con.attached.Wait()
	if err != nil {
		con.attached.Release()
	}
	return err
}

func (con *lxcConsole) ForceClose() error {
	if con.attached == nil {
		return fmt.Errorf("No process is found")
	}
	// This sends just signal. pty and pid are
	// closed by Wait()
	return con.attached.Kill()
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
	go func() {
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
	options.StdinFd = ftty.Fd()
	options.StdoutFd = ftty.Fd()
	options.StderrFd = ftty.Fd()
	options.ClearEnv = true

	pid, err := c.RunCommandNoWait([]string{"/bin/bash"}, options)
	if err != nil {
		con.lxc.log.WithError(err).Errorln("Failed to AttachShell")
		defer fpty.Close()
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
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
	options.StdinFd = ftty.Fd()
	options.StdoutFd = ftty.Fd()
	options.StderrFd = ftty.Fd()
	options.EscapeCharacter = '~'

	if err := c.Console(options); err != nil {
		con.lxc.log.Errorln(err)
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
