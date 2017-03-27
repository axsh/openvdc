package null

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
)

func init() {
	hypervisor.RegisterProvider("null", &NullHypervisorProvider{})
}

type NullHypervisorProvider struct {
}

func (n *NullHypervisorProvider) Name() string {
	return "null"
}

func (n *NullHypervisorProvider) CreateDriver(string) (hypervisor.HypervisorDriver, error) {
	return &NullHypervisorDriver{}, nil
}

type NullHypervisorDriver struct {
}

func (h *NullHypervisorDriver) Recover(instanceState model.InstanceState) error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("Recover")
	return nil
}

func (h *NullHypervisorDriver) StartInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("StartInstance")
	return nil
}

func (h *NullHypervisorDriver) StopInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("StopInstance")
	return nil
}

func (h *NullHypervisorDriver) CreateInstance(*model.Instance, model.ResourceTemplate) error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("CreateInstance")
	return nil
}

func (h *NullHypervisorDriver) DestroyInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("DestroyInstance")
	return nil
}

func (h *NullHypervisorDriver) RebootInstance() error {
	log.WithFields(log.Fields{"reboot": "null"}).Infoln("RebootInstance")
	return nil
}

func (h *NullHypervisorDriver) InstanceConsole() hypervisor.Console {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("InstanceConsole")
	return nil
}
