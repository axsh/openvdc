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
	serialConn net.Conn
	conChan    chan error
	fds        []*os.File
	tty        string
}

// It seems like this function should just send the command and wait for the output
// instead, this mimics the lxc behavior by simply wrapping the Attach command...
// this is strange to me (maybe consider using the existing ConsoleParam struct to pass the args?).
func (con *qemuConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args...)
}

func (con *qemuConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
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
	socket := con.qemu.machine.SerialSocketPath
	con.serialConn, err = net.Dial("unix", socket)
	con.conChan = make(chan error, 1)

	if err != nil {
		defer con.serialConn.Close()
		return nil, errors.Wrap(err, join("", "\nFailed to connect to socket ", socket))
	}

	b := make([]byte, 8192)
	if _, err := con.serialConn.Write([]byte("\n")); err != nil {
		defer con.serialConn.Close()
		log.WithError(err).Error(join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n"))
	}

	if _, err := con.serialConn.Read(b); err != nil {
		defer con.serialConn.Close()
		log.WithError(err).Error(join("", "\nFailed to read the qemu socket buffer from socket ", socket))
	}

	log.Info("Received prompt")
	if len(args) == 0 {
		err = con.attachShell(param, waitClosed)
	} else {
		err = con.execCommand(param, waitClosed, args...)
	}
	if err != nil {
		defer con.serialConn.Close()
		closeChan <- err
	}

	go func () {
		defer close(closeChan)
		waitClosed.Wait()
	}()
	return closeChan, nil
}

func (con *qemuConsole) execCommand(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup, args ...string) error {
	waitClosed.Add(1)
	gaCmd := NewQEMUCommand(args, true)
	gaResp, err := gaCmd.SendCommand(con.qemu.machine.AgentSocketPath)
	if err != nil {
		return err
	}
	param.Stdout.Write([]byte(gaResp.Stdout))
	defer func() {
		con.conChan <- nil
		waitClosed.Done()
	}()
	return nil
}

func (con *qemuConsole) attachShell(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup) error  {
	waitClosed.Add(1)
	go func() {
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			select {
			case err := <- con.conChan:
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
				_, err = con.serialConn.Write(b[0:n])
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
			con.serialConn.SetDeadline(time.Now().Add(time.Second))
			n, err := con.serialConn.Read(b)
			select {
			case err := <- con.conChan:
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
				fmt.Println(join("", string(b[0:n]), " - from socket"))

				if bytes.Contains(b[0:n], []byte{0x11}) {
					log.WithError(err).Info("Received exit from stdin")
					con.conChan <- nil
					// conChan <- errors.Wrap(err, "\nConsole exited by ctrl-q\n\n")
					// log.WithError(err).Info("Stdin set ctrl-q error string")
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
		con.serialConn.Close()
		con.serialConn = nil
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
