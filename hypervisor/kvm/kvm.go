// +build linux

package lxc

import (
	"fmt"

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

type KVMHypervisorProivder {
}

type KVMHypervisorDriver struct {
	hypervisor.Base
	template  *model.KvmTemplate
	imageName string
	hostName  string
	machine   *qemu.Machine
}

func (p *KVMHypervisorProivder) Name () string {
	return "kvm"
}

func init() {
}

func (p *KVMHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	return nil
}

func (p *KVMHypervisorProivder) CreateDrivder (instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDeriver, error) {
	return nil
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

func (d KVMHypervisorDriver) InstanceConsole() Console {
	return nil
}
