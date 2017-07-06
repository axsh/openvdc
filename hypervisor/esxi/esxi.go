// +build linux

package esxi

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

type BridgeType int

const (
	None BridgeType = iota // 0
	Linux
	OVS
)

var settings struct {
	EsxiUser	string
	EsxiPass	string
	EsxiIp		string
	EsxiInsecure	bool
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
}

func (p *EsxiHypervisorProvider) Name () string {
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
	settings.EsxiInsecure = sub.GetString("hypervisor.esxi-insecure")

	return nil
}

func (p *EsxiHypervisorProvider) CreateDriver (instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	EsxiTmpl, ok := template.(*model.EsxiTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.WmwareTemplate: %T, template")
	}

	//Create VM

	driver := &EsxiHypervisorDriver{
		Base: hypervisor.Base{
			Log: log.WithFields(log.Fields{"Hypervisor": "esxi", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: EsxiTmpl,
		//vm: v,
	}
	return driver, nil
}

func (d *EsxiHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func (d *EsxiHypervisorDriver) CreateInstance() error {
	return nil
}

func (d *EsxiHypervisorDriver) DestroyInstance() error {
	return nil
}

func (d *EsxiHypervisorDriver) StartInstance() error {
	return nil
}

func (d *EsxiHypervisorDriver) StopInstance() error {
	return nil
}

func (d EsxiHypervisorDriver) RebootInstance() error {
	return nil
}

func (d EsxiHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
