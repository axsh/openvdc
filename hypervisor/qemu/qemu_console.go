// +build linux

package qemu

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
)

type qemuConsole struct {
	qemu       *QEMUHypervisorDriver
	attached   *os.Process // this should be removed in favor of a channel as below
	socketConn net.Conn
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
	if con.socketConn, err = net.Dial("unix", con.socketPath); err != nil {
		return nil, errors.Wrap(err, join(" ", "\nFailed to connect to socket", con.socketPath))
	}

	if len(args) == 0 {
		b := make([]byte, 8192)
		if _, err = con.socketConn.Write([]byte("\n")); err != nil {
			return nil, errors.Wrap(err, join(" ", "\nFailed to write to the qemu socket", con.socketPath, "from the buffer\n\n"))
		}
		if _, err = con.socketConn.Read(b); err != nil {
			return nil, errors.Wrap(err, join(" ", "\nFailed to read the qemu socket buffer from socket", con.socketPath))
		}
		log.Info("Received prompt")
		if err = con.attachShell(param, waitClosed); err != nil {
			return nil, err
		}
	} else {
		defer con.socketConn.Close()
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

	if err = NewGuestAgentExecRequest(args, true).SendRequest(con.socketConn, &execResp); err != nil {
		return err
	}
	if err = NewGuestAgentExecStatusRequest(execResp.Return.(*ExecResponse).Pid).SendRequest(con.socketConn, &statusResp); err != nil {
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

func (con *qemuConsole) attachShell(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup) error {
	waitClosed.Add(1)
	go func() {
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			select {
			case err := <-con.conChan:
				con.conChan <- err
				break
			default:
				n, err := param.Stdin.Read(b)
				if err != nil {
					con.conChan <- errors.Wrap(err, "\nFailed to read from the from the console input buffer\n\n")
				}
				fmt.Println(join("", string(b[0:n]), " - from stdin"))

				if bytes.Contains(b[0:n], []byte{0x11}) {
					log.Info("Received exit from stdin")
					con.conChan <- errors.Wrap(err, "\nConsole exited by ctrl-q\n\n")
				}
				_, err = con.socketConn.Write(b[0:n])
				if err != nil {
					con.conChan <- errors.Wrap(err, "\nFailed to write to the qemu socket from the buffer\n\n")
				}
			}
		}
		defer waitClosed.Done()
	}()

	waitClosed.Add(1)
	go func() {
		b := make([]byte, 8192)
		for {
			con.socketConn.SetDeadline(time.Now().Add(time.Second))
			n, err := con.socketConn.Read(b)
			select {
			case err := <-con.conChan:
				if _, e := param.Stdout.Write([]byte{0x0A}); e != nil {
					con.conChan <- errors.Wrap(e, "\nFailed to write the linefeed character on exit\n\n")
				} else {
					con.conChan <- err
				}
				break
			default:
				if err != nil && !err.(net.Error).Timeout() {
					con.conChan <- errors.Wrap(err, "\nFailed to read the qemu socket buffer from socket ")
				}
				// exit on ctrl + q
				if bytes.Contains(b[0:n], []byte{0x11}) {
					con.conChan <- nil
				}
				_, err = param.Stdout.Write(b[0:n])
				if err != nil {
					con.conChan <- errors.Wrap(err, "\nFailed to write to the console stdout buffer\n\n")
				}
			}
		}
		defer waitClosed.Done()
	}()

	return nil
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	// this should return when the user escapes the console.
	defer func() {
		con.socketConn.Close()
		con.socketConn = nil
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
