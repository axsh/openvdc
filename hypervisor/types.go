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
	CreateDriver(instance *model.Instance, template model.ResourceTemplate) (HypervisorDriver, error)
	LoadConfig(viper *viper.Viper) error
}

type HypervisorDriver interface {
	CreateInstance() error
	DestroyInstance() error
	StartInstance() error
	StopInstance() error
	RebootInstance() error
	InstanceConsole() Console
}

type Closed error

type ConsoleWaitError interface {
	error
	ExitCode() int
}

type Console interface {
	Attach(param *ConsoleParam) (<-chan Closed, error)
	Exec(param *ConsoleParam, args []string) (<-chan Closed, error)
	Wait() error
	ForceClose() error
}

type PtyConsole interface {
	AttachPty(param *ConsoleParam, ptyreq *SSHPtyReq) (<-chan Closed, error)
	UpdateWindowSize(w, h uint32) error
}

type ConsoleParam struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Envs   map[string]string
}

func NewConsoleParam(stdin io.Reader, stdout, stderr io.Writer) *ConsoleParam {
	return &ConsoleParam{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Envs:   make(map[string]string),
	}
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
