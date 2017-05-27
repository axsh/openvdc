package null

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
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

func (n *NullHypervisorProvider) CreateDriver(instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	nullTempl, ok := template.(*model.NullTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.NullTemplate: %T", template)
	}
	driver := &NullHypervisorDriver{
		Base: hypervisor.Base{
			Log:      log.WithFields(log.Fields{"hypervisor": "null", "instance_id": instance.Id}),
			Instance: instance,
		},
		template: nullTempl,
	}
	return driver, nil
}

type NullHypervisorDriver struct {
	hypervisor.Base
	template *model.NullTemplate
}

func (h *NullHypervisorDriver) log() *log.Entry {
	return h.Base.Log
}

func (h *NullHypervisorDriver) Recover(instanceState model.InstanceState) error {
	log.WithFields(log.Fields{"hypervisor": "null"}).Infoln("Recover")
	return nil
}

func (h *NullHypervisorDriver) StartInstance() error {
	h.log().Infoln("StartInstance")
	if h.template.CrashStage == model.NullTemplate_START {
		return errors.Errorf("Emulate crash at StartInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) StopInstance() error {
	h.log().Infoln("StopInstance")
	if h.template.CrashStage == model.NullTemplate_STOP {
		return errors.Errorf("Emulate crash at StoptInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) CreateInstance() error {
	h.log().Infoln("CreateInstance")
	if h.template.CrashStage == model.NullTemplate_CREATE {
		return errors.Errorf("Emulate crash at CreateInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) DestroyInstance() error {
	h.log().Infoln("DestroyInstance")
	if h.template.CrashStage == model.NullTemplate_DESTROY {
		return errors.Errorf("Emulate crash at DestroyInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) RebootInstance() error {
	h.log().Infoln("RebootInstance")
	if h.template.CrashStage == model.NullTemplate_REBOOT {
		return errors.Errorf("Emulate crash at RebootInstance")
	}
	return nil
}

func (h *NullHypervisorDriver) InstanceConsole() hypervisor.Console {
	h.log().Infoln("InstanceConsole")
	return nil
}
