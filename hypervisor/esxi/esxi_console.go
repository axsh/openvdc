package esxi

import (
	"fmt"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/hypervisor/util"
	"io"
	"net"
	"bytes"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	_ "github.com/vmware/govmomi/govc/vm"
	_ "github.com/vmware/govmomi/govc/vm/guest"
)

type esxiConsole struct {
	util.SerialConnection

	esxi    *EsxiHypervisorDriver
	conChan chan error
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
		b := make([]byte, 8192)
		if _, err = con.SerialConn.Write([]byte("\n")); err != nil {
			return nil, errors.Wrap(err, "\nFailed to write to the serial connection from the buffer\n\n")
		}
		if _, err = con.SerialConn.Read(b); err != nil {
			return nil, errors.Wrap(err, "\nFailed to read the serial connection buffer")
		}
		con.AttachSerialConsole(param, waitClosed, con.conChan)
	} else {
		con.execCommand(param, waitClosed, args...)
	}

	go func() {
		waitClosed.Wait()
		defer close(closeChan)
	}()

	return closeChan, nil
}

func (con *esxiConsole) execCommand(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup, args ...string) {
	waitClosed.Add(1)
	waitError := &consoleWaitError{exitCode: 1}
	output := make(chan string)
	cmd := []string{"guest.run", fmt.Sprintf("-vm.path=[%s]%s/%s.vmx", settings.EsxiVmDatastore, con.esxi.vmName, con.esxi.vmName)}

	// the exec command from sshd includes /bin/sh -c here, which is not compatible with the guest.run command
	for _, arg := range strings.Split(args[2], " ") {
		cmd = append(cmd, arg)
	}

	rOut, wOut, err := os.Pipe()
	if err != nil {
		log.Info("failed os.Pipe for stdout")
	}
	stdout := os.Stdout // original stdout
	os.Stdout = wOut

	rErr, wErr, err := os.Pipe()
	if err != nil {
		log.Info("failed os.Pipe for stderr")
	}
	stderr := os.Stderr // original stderr
	os.Stderr = wErr

	waitError.exitCode = esxiRunCmd(cmd)
	waitClosed.Add(1)
	go func() {
		var buf bytes.Buffer
		if waitError.ExitCode() == 0 {
			io.Copy(&buf, rOut)
		} else {
			io.Copy(&buf, rErr)
		}
		output <- buf.String()
		defer waitClosed.Done()
	}()
	defer func() {
		con.conChan <- waitError
		waitClosed.Done()
	}()

	wOut.Close()
	wErr.Close()
	os.Stdout = stdout
	os.Stderr = stderr

	fmt.Fprint(param.Stdout, <- output)
}

func (con *esxiConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param)
}

func (con *esxiConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args...)
}

func (con *esxiConsole) Wait() error {
	defer func() {
		if con.SerialConn != nil {
			con.SerialConn.Close()
		}
	}()
	return <-con.conChan
}

func (con *esxiConsole) ForceClose() error {
	return nil
}

type consoleWaitError struct {
	exitCode int
	err      error
}

func (e *consoleWaitError) Error() string {
	return fmt.Sprintf("Process failed with %d", e.exitCode)
}

func (e *consoleWaitError) ExitCode() int {
	if e.exitCode != 0 {
		return 1
	}
	return 0
}
