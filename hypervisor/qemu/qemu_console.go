// +build linux

package qemu

import (
	"bufio"
	"bytes"
	"io"
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

func (console *qemuConsole) unixSocketConsole(stdin, stdout, stderr *os.File) error {
	socket := console.machine().Serial // if this socket is closed, we have no way to reopen it...yet

	in := bufio.NewReader(stdin)
	out := bufio.NewWriter(stdout)
	stdErr := bufio.NewWriter(stderr)

	connection, err := net.Dial("unix", socket)
	if err != nil {
		return errors.Wrap(err, join("", "\nFailed to connect to socket ", socket))
	}

	conChan := make(chan error)
	go func(r io.Reader, readErrChan chan error) {
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			n, err := r.Read(b)
			if err != nil {
				readErrChan <- errors.Wrap(err, join("", "\nFailed to read the qemu socket buffer from socket ", socket))
				break
			}

			if bytes.Contains(b[0:n], []byte{0x01}) {
				readErrChan <- errors.Wrap(err, "\nConsole exited by ctrl-a\n\n")
				break
			}

			_, err = out.Write(b[0:n]) //this err could probably be scoped to only this block.
			if err != nil {
				readErrChan <- errors.Wrap(err, "\nFailed to write to the console stdout buffer\n\n")
				break
			}

			_, err = stdErr.Write(b[0:n]) //this needs it's own listener to make a console. it is just a terminal as is.
			if err != nil {
				readErrChan <- errors.Wrap(err, "\nFailed to write to the console stderr buffer\n\n")
				break
			}
		}
	}(connection, conChan)

	go func(w io.Writer, writeErrChan chan error) {
		b := make([]byte, 8192)
		for {
			// err is for reads or writes shorter than the buffer size -- should probably retry in that case...
			n, err := in.Read(b)
			if err != nil {
				writeErrChan <- errors.Wrap(err, "\nFailed to read from the from the console input buffer\n\n")
				break
			}
			if bytes.Contains(b[0:n], []byte{0x11}) {
				writeErrChan <- errors.Wrap(err, "\nConsole exited by ctrl-q\n\n")
				break
			}
			_, err = w.Write(b[0:n])
			if err != nil {
				writeErrChan <- errors.Wrap(err, join("", "\nFailed to write to the qemu socket ", socket, " from the buffer\n\n"))
				break
			}
		}
	}(connection, conChan)

	console.conChan = conChan

	return nil
}

func (con *qemuConsole) pipeAttach(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	// if con.machine().State != RUNNING {
	// 	return nil, errors.New("kvm instance is not in a running state")
	// }

	fds := make([]*os.File, 6)
	closeAll := func() {
		for _, fd := range fds {
			fd.Close()
		}
	}

	var err error
	rIn, wIn, err := os.Pipe() // stdin
	if err != nil {
		return nil, errors.Wrap(err, "Failed os.Pipe for stdin")
	}
	fds = append(fds, rIn, wIn)
	con.fds = append(con.fds, wIn)
	// con.fds = append(con.fds, rIn)
	rOut, wOut, err := os.Pipe() // stdout
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stdout")
	}
	fds = append(fds, rOut, wOut)
	con.fds = append(con.fds, rOut)
	// con.fds = append(con.fds, wOut)
	rErr, wErr, err := os.Pipe() // stderr
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stderr")
	}
	fds = append(fds, rErr, wErr)
	con.fds = append(con.fds, rErr)
	// con.fds = append(con.fds, wErr)

	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed, 3)
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(wIn, param.Stdin)
		// _, err := io.Copy(rIn, param.Stdin)
		if err == nil {
			// param.Stdin was closed due to EOF so needs to send EOF to pipe as well
			wIn.Close()
		}
		// TODO: handle err is not nil case
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stdout, rOut)
		// _, err := io.Copy(param.Stdout, wOut)
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stderr, rErr)
		// _, err := io.Copy(param.Stderr, wErr)
		closeChan <- err
		defer waitClosed.Done()
	}()
	go func() {
		waitClosed.Wait()
		defer close(closeChan)
	}()

	if err := con.unixSocketConsole(rIn, wOut, wErr); err != nil {
		// if err := con.unixSocketConsole(wIn, rOut, rErr); err != nil {
		defer func() {
			closeAll()
			con.fds = []*os.File{}
		}()
		return nil, err
	}

	// Close file descriptors for child process.
	defer func() {
		rIn.Close()
		wOut.Close()
		wErr.Close()
		// wIn.Close()
		// rOut.Close()
		// rErr.Close()
	}()

	return closeChan, nil
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	// this should return nil when the user escapes the console.
	conChan := <-con.conChan
	if conChan != nil {
		return conChan
	}
	return nil
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
