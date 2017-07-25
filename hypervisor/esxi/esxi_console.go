package esxi

import(
	"github.com/axsh/openvdc/hypervisor"
)

type esxiConsole struct {
	esxi *EsxiHypervisorDriver
}

func (d *EsxiHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &esxiConsole{
		esxi: d,
	}
}

func (c *esxiConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return nil,nil
}

func (c *esxiConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return nil, nil
}

func (c *esxiConsole) Wait () error {
	return nil
}

func (c *esxiConsole) ForceClose() error {
	return nil
}

type consoleWaitError struct {
}

func (e *consoleWaitError) ExitCode() int {
	return 0
}
