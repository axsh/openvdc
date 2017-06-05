// +build linux

package kvm

import (
	// "fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
	qemu "github.com/quadrifoglio/go-qemu"
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

type KVMHypervisorProvider struct {
}

type KVMHypervisorDriver struct {
	hypervisor.Base
	template  *model.KvmTemplate
	imageName string
	hostName  string
	machine   qemu.Machine
}

func (p *KVMHypervisorProvider) Name () string {
	return "kvm"
}

func init() {
	hypervisor.RegisterProvider("kvm", &KVMHypervisorProvider{})
}

func (p *KVMHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	return nil
}

func (p *KVMHypervisorProvider) CreateDrivder (instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	kvmTmpl, ok := template.(*model.KvmTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.KvmTemplate: %T, template")
	}
	m := qemu.NewMachine(int(kvmTmpl.Vcpu), uint64(kvmTmpl.MemoryGb))
	driver := &KVMHypervisorDriver{
		Base: hypervisor.Base{
			Log: log.WithFields(log.Fields{"Hypervisor": "kvm", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: kvmTmpl,
		machine: m,
	}
	return driver, nil
}

func (d *KVMHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func (d *KVMHypervisorDriver) CreateInstance() error {
	return nil
}

func (d *KVMHypervisorDriver) DestroyInstance() error {
	return nil
}

func (d *KVMHypervisorDriver) StartInstance() error {
	return nil
}

func (d *KVMHypervisorDriver) StopInstance() error {
	return nil
}

func (d KVMHypervisorDriver) RebootInstance() error {
	return nil
}

func (d KVMHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
