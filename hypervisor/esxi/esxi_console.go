package esxi

import (
	"github.com/axsh/openvdc/hypervisor/util"
	"github.com/axsh/openvdc/hypervisor"
	"net"
	"strconv"
	"strings"
	"sync"

	// log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type esxiConsole struct {
	util.SerialConnection

	esxi       *EsxiHypervisorDriver
	conChan    chan error
}
func (d *EsxiHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &esxiConsole{
		esxi: d,
	}
}

func (c *esxiConsole) pipeAttach(param *hypervisor.ConsoleParam, args ...string) (<-chan hypervisor.Closed, error) {
	//TODO: check if machine is running {
	// return nil, errors.New("esxi instance is not in a running state")
	//}
	var err error
	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed)
	c.conChan = make(chan error)

	if len(args) == 0 {
		c.SerialConn, err = net.Dial("tcp", strings.Join([]string{settings.EsxiIp, strconv.Itoa(c.esxi.machine.SerialConsolePort)}, ":"))
		if err != nil {
			return nil, errors.Errorf("Unable to connect to %s on port %d", settings.EsxiIp, c.esxi.machine.SerialConsolePort)
		}
		go func() {
			waitClosed.Wait()
		}()
		c.AttachSerialConsole(param, waitClosed, c.conChan)
	} else {
		c.execCommand(param, waitClosed, args...)
	}

	defer close(closeChan)
	return closeChan, nil
}

func (c *esxiConsole) execCommand(param *hypervisor.ConsoleParam, waitDone *sync.WaitGroup, args ...string) {
	c.conChan <- nil
}

func (c *esxiConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return c.pipeAttach(param)
}

func (c *esxiConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return c.pipeAttach(param, args...)
}

func (c *esxiConsole) Wait() error {
	defer c.SerialConn.Close()
	return <-c.conChan
}

func (c *esxiConsole) ForceClose() error {
	return nil
}

type consoleWaitError struct {
}

func (e *consoleWaitError) ExitCode() int {
	return 0
}
