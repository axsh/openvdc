package null

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
)

func init() {
	hypervisor.RegisterProvider("null", &NullHypervisorProvider{})
}

type NullHypervisorProvider struct {
}

func (n *NullHypervisorProvider) SetName(string) {

}

func (n *NullHypervisorProvider) Name() string {
	return "null"
}

func (n *NullHypervisorProvider) CreateDriver() (hypervisor.HypervisorDriver, error) {
	return &NullHypervisorDriver{}, nil
}

type NullHypervisorDriver struct {
}

func (h *NullHypervisorDriver) StartInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("StartInstance")
	return nil
}

func (h *NullHypervisorDriver) StopInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("StopInstance")
	return nil
}

func (h *NullHypervisorDriver) CreateInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("CreateInstance")
	return nil
}

func (h *NullHypervisorDriver) DestroyInstance() error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("DestroyInstance")
	return nil
}

func (h *NullHypervisorDriver) InstanceConsole() error {
        log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("InstanceConsole")
        return nil
}
