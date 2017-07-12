// +build linux

package esxi

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	cli "github.com/vmware/govmomi/govc/cli"
	_ "github.com/vmware/govmomi/govc/datastore"
	_ "github.com/vmware/govmomi/govc/vm"
)

type BridgeType int

const (
	None BridgeType = iota // 0
	Linux
	OVS
)

var settings struct {
	EsxiUser        string
	EsxiPass        string
	EsxiIp          string
	EsxiInsecure    bool
	EsxiHostSshkey  string
	EsxiVmUser      string
        EsxiVmPass      string
	EsxiVmDatastore string
	EsxiUrl         string
	BridgeName      string
	BridgeType      BridgeType
}

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

type EsxiHypervisorProvider struct {
}

type EsxiHypervisorDriver struct {
	hypervisor.Base
	template  *model.EsxiTemplate
	imageName string
	hostName  string
	vmName    string
}

func (p *EsxiHypervisorProvider) Name() string {
	return "esxi"
}

func init() {
	hypervisor.RegisterProvider("esxi", &EsxiHypervisorProvider{})
	viper.SetDefault("hypervisor.esxi-insecure", true)
}

func (p *EsxiHypervisorProvider) LoadConfig(sub *viper.Viper) error {
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

	settings.EsxiUser = sub.GetString("hypervisor.esxi-user")
	settings.EsxiPass = sub.GetString("hypervisor.esxi-pass")
	settings.EsxiIp = sub.GetString("hypervisor.esxi-ip")
	settings.EsxiInsecure = sub.GetBool("hypervisor.esxi-insecure")
	settings.EsxiHostSshkey = sub.GetString("hypervisor.esxi-host-sshkey")
	settings.EsxiVmUser = sub.GetString("hypervisor.esxi-vm-user")
	settings.EsxiVmPass = sub.GetString("hypervisor.esxi-vm-pass")
	settings.EsxiVmDatastore = sub.GetString("hypervisor.esxi-vm-datastore")

	esxiInfo := fmt.Sprintf("%s:%s@%s", settings.EsxiUser, settings.EsxiPass, settings.EsxiIp)
	u, _ := url.Parse("https://")
	u.Path = path.Join(u.Path, esxiInfo, "sdk")
	settings.EsxiUrl = u.String()

	return nil
}

func (p *EsxiHypervisorProvider) CreateDriver(instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	EsxiTmpl, ok := template.(*model.EsxiTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.WmwareTemplate: %T, template")
	}

	driver := &EsxiHypervisorDriver{
		Base: hypervisor.Base{
			Log:      log.WithFields(log.Fields{"Hypervisor": "esxi", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: EsxiTmpl,
		vmName:   instance.GetId(),
	}

	return driver, nil
}

func (d *EsxiHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func esxiCmd(args ...string) {

	var a []string

	a = append(a, args[0])
	a = append(a, fmt.Sprintf("-dc=%s", "ha-datacenter"))
	a = append(a, fmt.Sprintf("-k=%s", "true"))
	a = append(a, fmt.Sprintf("-u=%s", settings.EsxiUrl))

	for i := 1; i < len(args); i++ {
		a = append(a, args[i])
	}

	cli.Run(a)
}

func (d *EsxiHypervisorDriver) CreateInstance() error {

	// Create new folder
	esxiCmd("datastore.mkdir", fmt.Sprintf("-ds=%s", "datastore2"), d.vmName)

	// Ssh into esxiHost and use "vmkfstools" to clone vmdk"
	vmkfstoolsCmd := fmt.Sprintf("vmkfstools -i /vmfs/volumes/%s/%s/%s.vmdk /vmfs/volumes/%s/%s/%s.vmdk -d thin",
		"datastore2", "Centos7", "Centos7", "datastore2", d.vmName, "Centos7") //Todo: Don't use hardcoded values
	cmd := exec.Command("ssh", "-i", settings.EsxiHostSshkey, "-o", "StrictHostKeyChecking=no", "-o", "LogLevel=quiet", "-o", "UserKnownHostsFile /dev/null", fmt.Sprintf("root@%s", settings.EsxiIp), vmkfstoolsCmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String()) //Todo: Return error in better way
		log.Fatal(err)
	}

	//Copy .vmx-file
	esxiCmd("datastore.cp", fmt.Sprintf("-ds=%s", "datastore2"), fmt.Sprintf("%s/%s.vmx", "Centos7", "Centos7"), fmt.Sprintf("%s/%s.vmx", d.vmName, d.vmName))

	//Register new VM
	esxiCmd("vm.register", fmt.Sprintf("-ds=%s", "datastore2"), fmt.Sprintf("%s/%s.vmx", d.vmName, d.vmName))

	//Rename VM
	esxiCmd("vm.change", fmt.Sprintf("-name=%s", d.vmName), fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))

	//Start VM
	esxiCmd("vm.power", "-on=true", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))

	return nil
}

func (d *EsxiHypervisorDriver) DestroyInstance() error {

	esxiCmd("vm.power", "-on=false", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))
	esxiCmd("vm.destroy", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))

	return nil
}

func (d *EsxiHypervisorDriver) StartInstance() error {
        esxiCmd("vm.power", "-on=true","-suspend=false", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))

	return nil
}

func (d *EsxiHypervisorDriver) StopInstance() error {
	//Suspend to save machine state
	esxiCmd("vm.power", "-suspend=true", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", "datastore2", d.vmName, d.vmName))

	return nil
}

func (d EsxiHypervisorDriver) RebootInstance() error {
	return nil
}

func (d EsxiHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
