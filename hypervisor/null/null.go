package null

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
)

func init() {
	hypervisor.RegisterProvider("null", &NullHypervisorProvider{})
}

type NullHypervisorProvider struct {
}

func (n *NullHypervisorProvider) Name() string {
	return "null"
}

func (n *NullHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	return nil
}

func (n *NullHypervisorProvider) CreateDriver(string) (hypervisor.HypervisorDriver, error) {
	return &NullHypervisorDriver{log: log.WithField("hypervisor", "null")}, nil
}

type NullHypervisorDriver struct {
	log                    *log.Entry
}

func (h *NullHypervisorDriver) StartInstance() error {
	h.log.Infoln("StartInstance")
	return nil
}

func (h *NullHypervisorDriver) StopInstance() error {
	h.log.Infoln("StopInstance")
	return nil
}

func (h *NullHypervisorDriver) CreateInstance(*model.Instance, model.ResourceTemplate) error {
	h.log.Infoln("CreateInstance")
	return nil
}

func (h *NullHypervisorDriver) DestroyInstance() error {
	h.log.Infoln("DestroyInstance")
	return nil
}

func (h *NullHypervisorDriver) RebootInstance() error {
	h.log.Infoln("RebootInstance")
	return nil
}

func (h *NullHypervisorDriver) InstanceConsole() hypervisor.Console {
	h.log.Infoln("InstanceConsole")
	return nil
}
