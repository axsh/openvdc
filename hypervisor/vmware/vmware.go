// +build linux

package vmware

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
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

type VmwareHypervisorProvider struct {
}

type WmwareHypervisorDriver struct {
	hypervisor.Base
	template  *model.WmwareTemplate
	imageName string
	hostName  string
}

func (p *WmwareHypervisorProvider) Name () string {
	return "wmware"
}

func init() {
	hypervisor.RegisterProvider("wmware", &WmwareHypervisorProvider{})
}

func (p *WmwareHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	return nil
}

func (p *WmwareHypervisorProvider) CreateDriver (instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	WmwareTmpl, ok := template.(*model.WmwareTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.WmwareTemplate: %T, template")
	}

	//Create VM

	driver := &WmwareHypervisorDriver{
		Base: hypervisor.Base{
			Log: log.WithFields(log.Fields{"Hypervisor": "vmware", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: vmwareTmpl,
		//vm: v,
	}
	return driver, nil
}

func (d *WmwareHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func (d *WmwareHypervisorDriver) CreateInstance() error {
	return nil
}

func (d *WmwareHypervisorDriver) DestroyInstance() error {
	return nil
}

func (d *WmwareHypervisorDriver) StartInstance() error {
	return nil
}

func (d *WmwareHypervisorDriver) StopInstance() error {
	return nil
}

func (d WmwareHypervisorDriver) RebootInstance() error {
	return nil
}

func (d WmwareHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
