// +build linux

package lxc

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

var ConfigFile string

const (
	ScriptPath      = "/etc/lxc/"
	LinuxUpScript   = "linux-bridge-up.sh"
	LinuxDownScript = "linux-bridge-down.sh"
	OvsUpScript     = "ovs-up.sh"
	OvsDownScript   = "ovs-down.sh"
)

type NetworkInterface struct {
	Type string
	BridgeName string
	Ipv4Addr   string
	MacAddr    string
}

var interfaces []NetworkInterface


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

func modifyConf(scriptUp string, scriptDown string) {
	f, err := ioutil.ReadFile(ConfigFile)

	if err != nil {
		log.Fatalf("Failed loading lxc default.conf: ", err)
	}

	// Give a warning if script doesn't exist.
	_, err = ioutil.ReadFile(scriptUp)
	if err != nil {
		log.Warn("File not found: "+scriptUp, err)
	}

	// Give a warning if script doesn't exist.
	_, err = ioutil.ReadFile(scriptDown)
	if err != nil {
		log.Warn("File not found: " + scriptDown, err)
	}

	scripts := "lxc.network.script.up = " + ScriptPath + scriptUp + "\nlxc.network.script.down = " + ScriptPath + scriptDown

	cf := cleanConfigFile(string(f))

	result := strings.Join([]string{cf, scripts}, "")
	err = ioutil.WriteFile(ConfigFile, []byte(result), 0644)
	if err != nil {
		log.Fatalln(err)
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

func handleNetworkSettings(nwi NetworkInterface) {

	switch nwi.Type {
	case "linux":
		modifyConf(LinuxUpScript, LinuxDownScript)
	case "ovs":
		modifyConf(OvsUpScript, OvsDownScript)
	default:
	}
}

func (d *LXCHypervisorDriver) CreateInstance(i *model.Instance, in model.ResourceTemplate) error {

	lxcTmpl, ok := in.(*model.LxcTemplate)

	if !ok {

		log.Fatal("BUGON: Unsupported model type")

	}

	c, err := lxc.NewContainer(d.name, d.lxcpath)

	ConfigFile = c.ConfigFileName()

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

		interfaces = append(interfaces,
               		NetworkInterface{
                        	BridgeName: "br0",
                        	Type: i.GetType(),
				Ipv4Addr: i.GetIpv4Addr(),
        			MacAddr: i.GetMacaddr(),
                	},
        	)
	}

        handleNetworkSettings(interfaces[0])

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

func (d *LXCHypervisorDriver) InstanceConsole() error {

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	if err != nil {
		d.log.Errorln(err)
		return err
	}

	if c.State() != lxc.RUNNING {
		d.log.Errorf("lxc-container can not perform console")
		return fmt.Errorf("lxc-container can not perform console")
	}

	var options = lxc.ConsoleOptions{
		Tty:             -1,
		StdinFd:         os.Stdin.Fd(), //These should probably be changed.
		StdoutFd:        os.Stdout.Fd(),
		StderrFd:        os.Stderr.Fd(),
		EscapeCharacter: 'a',
	}

	if err := c.Console(options); err != nil {
		d.log.Errorln(err)
		return err
	}
	return nil
}
