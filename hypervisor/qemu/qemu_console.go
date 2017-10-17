// +build linux

package qemu

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/hypervisor/util"
	"github.com/pkg/errors"
)

type qemuConsole struct {
	util.SerialConnection

	qemu       *QEMUHypervisorDriver
	attached   *os.Process // this should be removed in favor of a channel as below
	socketPath string
	conChan    chan error
	fds        []*os.File
	tty        string
}

// It seems like this function should just send the command and wait for the output
// instead, this mimics the lxc behavior by simply wrapping the Attach command...
// this is strange to me (maybe consider using the existing ConsoleParam struct to pass the args?).
func (con *qemuConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	con.socketPath = con.qemu.machine.AgentSocketPath
	return con.pipeAttach(param, args...)
}

func (con *qemuConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	con.socketPath = con.qemu.machine.SerialSocketPath
	return con.pipeAttach(param)
}

func join(separator string, args ...string) string {
	// a wrapper for strings.Join to avoid writing the syntax for a string array every time
	// (in most simple cases, the strings.Join method is the most efficient way to join strings
	// http://herman.asia/efficient-string-concatenation-in-go, and its very convenient to use!)
	return strings.Join(args, separator)
}

func (con *qemuConsole) pipeAttach(param *hypervisor.ConsoleParam, args ...string) (<-chan hypervisor.Closed, error) {
	if !con.qemu.machine.HavePrompt() {
		return nil, errors.New("kvm instance is not in a running state")
	}

	var err error
	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed)
	con.conChan = make(chan error, 1)

	if con.SerialConn, err = net.Dial("unix", con.socketPath); err != nil {
		return nil, errors.Wrap(err, join(" ", "\nFailed to connect to socket", con.socketPath))
	}

	if len(args) == 0 {
		b := make([]byte, 8192)
		if _, err = con.SerialConn.Write([]byte("\n")); err != nil {
			return nil, errors.Wrap(err, join(" ", "\nFailed to write to the qemu socket", con.socketPath, "from the buffer\n\n"))
		}
		if _, err = con.SerialConn.Read(b); err != nil {
			return nil, errors.Wrap(err, join(" ", "\nFailed to read the qemu socket buffer from socket", con.socketPath))
		}
		log.Info("Received prompt")
		if err = con.AttachSerialConsole(param, waitClosed, con.conChan); err != nil {
			return nil, err
		}
	} else {
		defer con.SerialConn.Close()
		if err = con.execCommand(param, waitClosed, args...); err != nil {
			return nil, err
		}
	}
	go func() {
		defer close(closeChan)
		waitClosed.Wait()
	}()
	return closeChan, nil
}

func (con *qemuConsole) execCommand(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup, args ...string) error {
	waitClosed.Add(1)
	var err error
	var execResp GuestAgentResponse
	var statusResp GuestAgentResponse

	if err = NewGuestAgentExecRequest(args, true).SendRequest(con.SerialConn, &execResp); err != nil {
		return err
	}
	if err = NewGuestAgentExecStatusRequest(execResp.Return.(*ExecResponse).Pid).SendRequest(con.SerialConn, &statusResp); err != nil {
		return err
	}

	resp := statusResp.Return.(*ExecStatusResponse)
	cmdStderr := Base64toUTF8(resp.Stderr)
	cmdStdout := Base64toUTF8(resp.Stdout)
	defer func() {
		con.conChan <- &consoleWaitError{
			ExecStatusResponse: resp,
		}
		waitClosed.Done()
	}()

	if resp.Exitcode != 0 {
		param.Stdout.Write([]byte(cmdStderr))
	} else {
		param.Stdout.Write([]byte(cmdStdout))
	}
	return nil
}

type consoleWaitError struct {
	*ExecStatusResponse
}

func (c *consoleWaitError) Error() string {
	return fmt.Sprintf("Process failed with %d", c.Exitcode)
}

func (c *consoleWaitError) ExitCode() int {
	if c.Exitcode != 0 {
		return 1
	}
	return 0
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	// this should return when the user escapes the console.
	defer func() {
		con.SerialConn.Close()
		con.SerialConn = nil
		con.socketPath = ""
		con.fds = nil
		con.attached = nil
		con.conChan = nil
		con.tty = ""
		log.Info("Closed kvm console session")
	}()

	return <-con.conChan
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
