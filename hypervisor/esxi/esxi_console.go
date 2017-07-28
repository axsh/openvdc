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

func (con *esxiConsole) pipeAttach(param *hypervisor.ConsoleParam, args ...string) (<-chan hypervisor.Closed, error) {
	//TODO: check if machine is running {
	// return nil, errors.New("esxi instance is not in a running state")
	//}
	var err error
	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed)
	con.conChan = make(chan error)

	if len(args) == 0 {
		con.SerialConn, err = net.Dial("tcp", strings.Join([]string{settings.EsxiIp, strconv.Itoa(con.esxi.machine.SerialConsolePort)}, ":"))
		if err != nil {
			return nil, errors.Errorf("Unable to connect to %s on port %d", settings.EsxiIp, con.esxi.machine.SerialConsolePort)
		}
		go func() {
			waitClosed.Wait()
		}()
		con.AttachSerialConsole(param, waitClosed, con.conChan)
	} else {
		con.execCommand(param, waitClosed, args...)
	}

	defer close(closeChan)
	return closeChan, nil
}

func (con *esxiConsole) execCommand(param *hypervisor.ConsoleParam, waitDone *sync.WaitGroup, args ...string) {
	con.conChan <- nil
}

func (con *esxiConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param)
}

func (con *esxiConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args...)
}

func (con *esxiConsole) Wait() error {
	defer con.SerialConn.Close()
	return <-con.conChan
}

func (con *esxiConsole) ForceClose() error {
	return nil
}

type consoleWaitError struct {
}

func (e *consoleWaitError) ExitCode() int {
	return 0
}
