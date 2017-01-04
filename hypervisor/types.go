package hypervisor

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
)

type HypervisorProvider interface {
	Name() string
	CreateDriver(string) (HypervisorDriver, error)
}

type HypervisorDriver interface {
	CreateInstance(*model.Instance, model.ResourceTemplate) error
	DestroyInstance() error
	StartInstance() error
	StopInstance() error
	RebootInstance() error
	InstanceConsole() error
}

var (
	hypervisorProviders = make(map[string]HypervisorProvider)
)

func RegisterProvider(name string, p HypervisorProvider) error {
	if _, exists := hypervisorProviders[name]; exists {
		return fmt.Errorf("Duplicated hypervisor provider registration: %s", name)
	}
	hypervisorProviders[name] = p
	log.Infof("Registered hypervisor provider: %s\n", name)
	return nil
}

func FindProvider(name string) (p HypervisorProvider, ok bool) {
	p, ok = hypervisorProviders[name]
	return
}
