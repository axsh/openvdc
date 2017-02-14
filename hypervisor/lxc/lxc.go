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
	//for i, _ := range interfaces {
	newSettings = updateSettings(interfaces[0], newSettings)
	//}

	result := strings.Join([]string{cf, newSettings}, "")
	err = ioutil.WriteFile(LxcConfigFile, []byte(result), 0644)
	if err != nil {
		log.Fatalln("Couldn't save lxc config file", err)
	}
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

	var output string

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

		log.Fatal("BUGON: Unsupported model type")

	}

	c, err := lxc.NewContainer(d.name, d.lxcpath)
	LxcConfigFile = c.ConfigFileName()

	ContainerName = d.name

	if err != nil {

		d.log.Errorln(err)

		return err

	}

	d.log.Infoln("Creating lxc-container...")

	if err := c.Create(d.template); err != nil {

		d.log.Errorln(err)

		return err

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

	modifyConf()

	interfaces = nil

	return nil

}

func loadConfigFile() {
	viper.SetConfigName("executor")
	viper.AddConfigPath("/etc/openvdc/")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to load config %s: %v", viper.ConfigFileUsed(), err)
	}

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
