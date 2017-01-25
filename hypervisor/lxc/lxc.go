// +build linux

package lxc

import (
	"io"
	"os"
	"time"
	"fmt"
	"sync"
	log "github.com/Sirupsen/logrus"

  "github.com/kr/pty"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
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
	if ok := c.Wait(lxc.RUNNING, 30 * time.Second); !ok {
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
	if ok := c.Wait(lxc.STOPPED, 30 * time.Second); !ok {
		d.log.Errorln("Failed or timedout to wait for STOPPED")
		return fmt.Errorf("Failed or timedout to wait for STOPPED")
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

func (con *lxcConsole) attachShell(c *lxc.Container, stdin io.Reader, stdout, stderr io.Writer) error {
	var wg sync.WaitGroup
	fpty, ftty, err := pty.Open()
	if err != nil {
		return errors.Wrapf(err, "Failed to open tty")
	}
	defer ftty.Close()
	defer fpty.Close()

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

	options := lxc.DefaultAttachOptions
	options.StdinFd	= ftty.Fd()
	options.StdoutFd = ftty.Fd()
	options.StderrFd = ftty.Fd()
	options.ClearEnv = true

	if err := c.AttachShell(options); err != nil {
		con.lxc.log.WithError(err).Errorln("Failed to AttachShell")
		return err
	}
	wg.Wait()
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
		con.lxc.log.Errorln(err)
		return err
	}
	return nil
}
