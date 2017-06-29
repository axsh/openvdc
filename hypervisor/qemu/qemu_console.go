// +build linux

package qemu

import (
	"bytes"
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
	qemu     *QEMUHypervisorDriver
	attached *os.Process // this should be removed in favor of a channel as below
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
	// feel free to uncomment this if we ever add states for kvm
	// if con.machine().State != RUNNING {
	// 	return nil, errors.New("kvm instance is not in a running state")
	// }

	socket := con.machine().Serial // if this socket is closed, we have no way to reopen it...yet
	connection, err := net.Dial("unix", socket)
	if err != nil {
		return nil, errors.Wrap(err, join("", "\nFailed to connect to socket ", socket))
	}

	conChan := make(chan error)
	con.conChan = conChan
	errorString := "" // a channel might be a better way to do this, but this should still be ok

	waitClosed := new(sync.WaitGroup)
	waitClosed.Add(1)
	go func(errorString *string) {
		defer waitClosed.Done()
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			n, err := param.Stdin.Read(b)
			if err != nil {
				*errorString = "\nFailed to read from the from the console input buffer\n\n"
				conChan <- errors.Wrap(err, *errorString)
				break
			}

			//keylogger for testing
			// fmt.Println(join("", string(b[0:n]), " - from stdin"))

			//break loop if there is an error in the other goroutine
			if *errorString != "" {
				break
			}

			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.WithError(err).Info("Received exit from stdin")
				*errorString = "\nConsole exited by ctrl-q\n\n"
				conChan <- errors.Wrap(err, *errorString)
				// log.WithError(err).Info("Stdin set ctrl-q error string")
				break
			}
			_, err = connection.Write(b[0:n])
			if err != nil {
				*errorString = join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n")
				conChan <- errors.Wrap(err, *errorString)
				break
			}
		}
	}(&errorString)

	waitClosed.Add(1)
	go func(errorString *string) {
		defer waitClosed.Done()
		b := make([]byte, 8192)
		for {
			// set a timeout for the read call so that it is essentially non-blocking
			// increase this delay or change the timer method if performance is a problem.
			connection.SetDeadline(time.Now().Add(time.Second))
			n, err := connection.Read(b)
			if err != nil && !err.(net.Error).Timeout() {
				*errorString = join("", "\nFailed to read the qemu socket buffer from socket ", socket)
				conChan <- errors.Wrap(err, *errorString)
				break
			}

			//keylogger for testing
			// fmt.Println(join("", string(b[0:n]), " - from socket"))

			//break loop if there is an error in the other goroutine
			if *errorString != "" {
				break
			}

			_, err = param.Stdout.Write(b[0:n])
			if err != nil {
				*errorString = "\nFailed to write to the console stdout buffer\n\n"
				conChan <- errors.Wrap(err, *errorString)
				break
			}
		}
	}(&errorString)

	// waitClosed.Add(1)
	// go func() {
	// 	b := make([]byte, 8192)
	// 	for {
	// 		n, err := // something like connection.Stderr.Read(b) is required here to get full console functionality
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
	}()

	return make(chan hypervisor.Closed, 3), nil // as far as I can tell, this channel never gets used anywhere, but I am leaving it just in case (at any rate the only call I can find at line 169 of sshd.go is ignoring it)
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	// this should return when the user escapes the console.
	defer func() {
		con.fds = nil
		con.attached = nil
		con.conChan = nil
		con.tty = ""
		log.Info("Closed kvm console session")
	}()

	c := <-con.conChan
	return c
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
