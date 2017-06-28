// +build linux

package qemu

import (
	"bytes"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
)

type qemuConsole struct {
	qemu     *QEMUHypervisorDriver
	attached *os.Process // I think this should be removed in favor of a channel as below
	conChan  chan error
	fds      []*os.File
	tty      string
}

func (con *qemuConsole) machine() *Machine {
	return con.qemu.machine
}

func (con *qemuConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args)
}

func (con *qemuConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, []string{"/bin/bash"})
}

func join(separator string, args ...string) string {
	// a wrapper for strings.Join to avoid writing the syntax for a string array every time
	// (in most simple cases, the strings.Join method is the most efficient way to join strings -- http://herman.asia/efficient-string-concatenation-in-go, and its very convenient to use!)
	return strings.Join(args, separator)
}

func (con *qemuConsole) pipeAttach(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	// if con.machine().State != RUNNING {
	// 	return nil, errors.New("kvm instance is not in a running state")
	// }

	socket := con.machine().Serial // if this socket is closed, we have no way to reopen it...yet
	connection, connErr := net.Dial("unix", socket)
	if connErr != nil {
		return nil, errors.Wrap(connErr, join("", "\nFailed to connect to socket ", socket))
	}

	var errorString string
	var err error
	conChan := make(chan error)
	con.conChan = conChan
	exitChan := make(chan string)

	waitClosed := new(sync.WaitGroup)
	waitClosed.Add(1)
	go func() {
		b := make([]byte, 8192)
		for {
			n, err := param.Stdin.Read(b)
			if err != nil {
				errorString = "\nFailed to read from the from the console input buffer\n\n"
				break
			}
			if bytes.Contains(b[0:n], []byte{0x11}) {
				exitChan <- "exit"
				errorString = "\nConsole exited by ctrl-q\n\n"
				break
			}
			_, err = connection.Write(b[0:n])
			if err != nil {
				errorString = join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n")
				break
			}
		}
		waitClosed.Done()
	}()

	waitClosed.Add(1)
	go func() {
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
	loop:
		for {
			select {
			case exit := <-exitChan:
				if exit == "exit" {
					break loop
				}
			default:
			}
			n, err := connection.Read(b)
			if err != nil {
				errorString = join("", "\nFailed to read the qemu socket buffer from socket ", socket)
				break
			}
			_, err = param.Stdout.Write(b[0:n])
			if err != nil {
				errorString = "\nFailed to write to the console stdout buffer\n\n"
				break
			}
		}
		waitClosed.Done()
	}()

	// waitClosed.Add(1)
	// go func() {
	// 	b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
	// 	for {
	// 		n, err := // something akin to connection.Stderr.Read(b) is required here to get full console functionality
	// 		if err != nil {
	// 			conChan <- errors.Wrap(err, join("", "\nFailed to read the qemu socket buffer from socket ", socket))
	// 			break
	// 		}
	// 		_, err = param.Stderr.Write(b[0:n])
	// 		if err != nil {
	// 			conChan <- errors.Wrap(err, "\nFailed to write to the console stderr buffer\n\n")
	// 			break
	// 		}
	// 	}
	// 	waitClosed.Done()
	// }()

	go func() {
		waitClosed.Wait()
		conChan <- errors.Wrap(err, errorString)
	}()

	return make(chan hypervisor.Closed), nil // as far as I can tell, this channel never gets used anywhere, but I am leaving it just in case (at any rate the only call I can find at line 169 of sshd.go is ignoring it)
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	// this should return when the user escapes the console.
	return <-con.conChan
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
