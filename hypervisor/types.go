package hypervisor

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
)

type HypervisorProvider interface {
	Name() string
	CreateDriver(instanceID string) (HypervisorDriver, error)
	LoadConfig(viper *viper.Viper) error
}

type HypervisorDriver interface {
	CreateInstance(*model.Instance, model.ResourceTemplate) error
	DestroyInstance() error
	StartInstance() error
	StopInstance() error
	InstanceConsole() Console
}

type Console interface {
	Attach(stdin io.Reader, stdout, stderr io.Writer) error
	Wait() error
	ForceClose() error
}

type PtyConsole interface {
	AttachPty(stdin io.Reader, stdout, stderr io.Writer, ptyreq *SSHPtyReq) error
}

// Compatible with "type ptyRequestMsg struct" in golang.org/x/crypto/ssh/session.go
type SSHPtyReq struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
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
