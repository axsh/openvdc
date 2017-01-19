// +build linux

package lxc

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"strconv"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/cmd/openvdc/constants"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

var LxcConfigFile string

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
	TapName    string
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

func modifyConf() {
	f, err := ioutil.ReadFile(LxcConfigFile)

	if err != nil {
		log.Fatalf("Failed loading lxc default.conf: ", err)
	}

	cf := cleanConfigFile(string(f))

	var newSettings string
	for i, _ := range interfaces {
		newSettings = updateSettings(interfaces[i], newSettings)		
	}

	result := strings.Join([]string{cf, newSettings}, "")
	err = ioutil.WriteFile(LxcConfigFile, []byte(result), 0644)
	if err != nil {
		log.Fatalln("Couldn't save lxc config file", err)
	}
}

func updateSettings(nwi NetworkInterface, input string) string {
	output := input + "\n" 

	if nwi.BridgeName != "" {
		output += fmt.Sprintf("#---- %s ----\n", nwi.BridgeName)
	}

	output += fmt.Sprintf("lxc.network.veth.pair=%s\n", nwi.TapName)

	if nwi.Ipv4Addr != "" {
		output += fmt.Sprintf("lxc.network.ipv4=%s\n", nwi.Ipv4Addr)
	}

	if nwi.MacAddr != "" {
                output += fmt.Sprintf("lxc.network.hwaddr=%s\n", nwi.MacAddr)
        }


	switch nwi.Type {
        case constants.BRIDGE_LINUX:
                output += fmt.Sprintf("lxc.network.script.up=%s\n", ScriptPath + LinuxUpScript)
		output += fmt.Sprintf("lxc.network.script.down=%s\n", ScriptPath + LinuxDownScript)
        case constants.BRIDGE_OVS:
                output += fmt.Sprintf("lxc.network.script.up=%s\n", ScriptPath + OvsUpScript)
                output += fmt.Sprintf("lxc.network.script.down=%s\n", ScriptPath + OvsDownScript)
        default:
		log.Fatalf("Unrecognized bridge type.")	
        }
	
	return output
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

func checkScript(script string) {
	_, err := os.Stat(ScriptPath+script)
	
        if err != nil {
                log.Infoln("Script not found.", err)
		createScript(ScriptPath+script)
        }
}

func createScript(script string) {
	log.Infoln("Creating script '%s'...", script) 
	f, err := os.Create(script)
    	
	if err != nil {
		log.Fatalf("Failed creating script.", err)
	}

	err = os.Chmod(script, 666)
	if err != nil {
     		log.Fatalf("Failed to set file permissions.", err)
 	}

	defer f.Close()
}


func (d *LXCHypervisorDriver) CreateInstance(i *model.Instance, in model.ResourceTemplate) error {

	lxcTmpl, ok := in.(*model.LxcTemplate)

	if !ok {

		log.Fatal("BUGON: Unsupported model type")

	}

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	LxcConfigFile = c.ConfigFileName()

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

	for x, i := range lxcTmpl.GetInterfaces() {

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
				TapName: d.name + "-" + strconv.Itoa(x),
                	},
        	)
	}
	
	checkScript(LinuxUpScript)
	checkScript(LinuxDownScript)
	checkScript(OvsUpScript)
	checkScript(OvsDownScript)

        modifyConf()

	interfaces = nil

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
		resetConfigFile()
		return err
	}

	d.log.Infoln("Waiting for lxc-container to become RUNNING")
	if ok := c.Wait(lxc.RUNNING, 30*time.Second); !ok {
		d.log.Errorln("Failed or timedout to wait for RUNNING")
		resetConfigFile()
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
