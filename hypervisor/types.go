package hypervisor

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
)

type ContainerState int

const (
	ContainerState_STOPPED ContainerState = iota + 1
	ContainerState_STARTING
	ContainerState_RUNNING
	ContainerState_STOPPING
	ContainerState_ABORTING
	ContainerState_NONE
)

type HypervisorProvider interface {
	Name() string
	CreateDriver(instanceID string) (HypervisorDriver, error)
}

type HypervisorDriver interface {
	CreateInstance(*model.Instance, model.ResourceTemplate) error
	DestroyInstance() error
	StartInstance() error
	StopInstance() error
	RebootInstance() error
	InstanceConsole() Console
	GetContainerState(*model.Instance) (ContainerState, error)
}

type Console interface {
	Attach(stdin io.Reader, stdout, stderr io.Writer) error
	Wait() error
	ForceClose() error
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
