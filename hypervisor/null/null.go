package null

import (
	"fmt"

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
	log        *log.Entry
	CrashStage model.NullTemplate_CrashStage
}

func (h *NullHypervisorDriver) StartInstance() error {
	h.log.Infoln("StartInstance")
	if h.CrashStage == model.NullTemplate_START {
		return fmt.Errorf("Emulate crash at StartInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) StopInstance() error {
	h.log.Infoln("StopInstance")
	if h.CrashStage == model.NullTemplate_STOP {
		return fmt.Errorf("Emulate crash at StoptInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) CreateInstance(*model.Instance, model.ResourceTemplate) error {
	h.log.Infoln("CreateInstance")
	if h.CrashStage == model.NullTemplate_CREATE {
		return fmt.Errorf("Emulate crash at CreateInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) DestroyInstance() error {
	h.log.Infoln("DestroyInstance")
	if h.CrashStage == model.NullTemplate_DESTROY {
		return fmt.Errorf("Emulate crash at DestroyInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) RebootInstance() error {
	h.log.Infoln("RebootInstance")
	if h.CrashStage == model.NullTemplate_REBOOT {
		return fmt.Errorf("Emulate crash at RebootInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) InstanceConsole() hypervisor.Console {
	h.log.Infoln("InstanceConsole")
	return nil
}
