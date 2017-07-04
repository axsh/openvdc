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
	qemu     *QEMUHypervisorDriver
	attached *os.Process // this should be removed in favor of a channel as below
	conChan  chan error
	fds      []*os.File
	tty      string
}

// It seems like this function should just send the command and wait for the output
// instead, this mimics the lxc behavior by simply wrapping the Attach command...
// this is strange to me (maybe consider using the existing ConsoleParam struct to pass the args?).
func (con *qemuConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, args...)
}

func (con *qemuConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return con.pipeAttach(param, "/bin/bash")
}

func join(separator string, args ...string) string {
	// a wrapper for strings.Join to avoid writing the syntax for a string array every time
	// (in most simple cases, the strings.Join method is the most efficient way to join strings -- http://herman.asia/efficient-string-concatenation-in-go, and its very convenient to use!)
	return strings.Join(args, separator)
}

func (con *qemuConsole) pipeAttach(param *hypervisor.ConsoleParam, args ...string) (<-chan hypervisor.Closed, error) {
	// feel free to uncomment this if we ever add states for kvm
	// if con.qemu.machine.State != RUNNING {
	// 	return nil, errors.New("kvm instance is not in a running state")
	// }

	socket := con.qemu.machine.Serial // if this socket is closed, we have no way to reopen it...yet
	connection, err := net.Dial("unix", socket)
	if err != nil {
		return nil, errors.Wrap(err, join("", "\nFailed to connect to socket ", socket))
	}

	b := make([]byte, 8192)
	if _, err := connection.Write([]byte("\n")); err != nil {
		log.Warn(join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n"))
		log.Info(join(" ", args...))
	}
	for {
		n, err := connection.Read(b)
		if err != nil {
			log.WithError(err).Error(join("", "\nFailed to read the qemu socket buffer from socket ", socket))
			break
		}
		if bytes.Contains(b[0:n], []byte{0x5D}) || bytes.Contains(b[0:n], []byte{0x3A}) { // check for "]" or ":" -- some sort of os.CmdPrompt() would be better, but this should be temporary.
			log.Info("Received prompt")
			break
		}
	}

	if _, err := connection.Write([]byte(join(join(" ", args...), "\n"))); err != nil {
		log.Warn(join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n"))
		log.Info(join(" ", args...))
	}
	conChan := make(chan error, 1)
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
			fmt.Println(join("", string(b[0:n]), " - from stdin"))

			//break loop if there is an error in the other goroutine
			if *errorString != "" {
				break
			}

			// if bytes.Contains(b[0:n], []byte{0x11}) {
			// 	log.Info("Received exit from stdin")
			// 	*errorString = "\nConsole exited by ctrl-q\n\n"
			// 	conChan <- errors.Wrap(err, *errorString)
			// 	// log.WithError(err).Info("Stdin set ctrl-q error string")
			// 	break
			// }
			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.Info("Received exit from stdin")
				*errorString = "\nConsole exited by ctrl-q\n\n"
				conChan <- errors.Wrap(err, *errorString)
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
			fmt.Println(join("", string(b[0:n]), " - from socket"))

			//break loop if there is an error in the other goroutine
			if *errorString != "" {
				if _, err := param.Stdout.Write([]byte{0x0A}); err != nil {
					*errorString = "\nFailed to write the linefeed character on exit\n\n"
					conChan <- errors.Wrap(err, *errorString)
				}
				break
			}

			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.WithError(err).Info("Received exit from stdin")
				*errorString = "\nConsole exited by ctrl-q\n\n"
				conChan <- errors.Wrap(err, *errorString)
				// log.WithError(err).Info("Stdin set ctrl-q error string")
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
		defer connection.Close()
		waitClosed.Wait()
	}()

	return make(chan hypervisor.Closed), nil // as far as I can tell, this channel never gets used anywhere, but I am leaving it just in case (at any rate the only calls I can find on lines 169 and 250 of sshd.go ignore it)
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

	return <-con.conChan
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
