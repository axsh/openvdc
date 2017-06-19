// +build linux

package qemu

import (
	"bufio"
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
	attached *os.Process
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
	//stdErr := bufio.NewReader(stderr)

	l, err := net.Listen("unix", socket)
	if err != nil {
		return errors.Wrap(err, join("", "Failed to listen to socket ", socket))
	}

	s, err := net.Dial("unix", socket)
	if err != nil {
		return errors.Wrap(err, join("", "Failed to send to socket ", socket))
	}
	go func(l net.Listener) error {
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			r, err := l.Accept()
			if err != nil {
				return errors.Wrap(err, join("Failed to accept the listening connection on socket ", socket))
			}
			n, err := r.Read(b)
			if err != nil {
				// these errors should be passed somewhere in channels...right now they are not handled anywhere.
				return errors.Wrap(err, join("", "Failed to read the qemu socket buffer from socket ", socket))
			}
			_, err = out.Write(b[0:n]) // err is for short writes -- should probably retry in that case...
			//_, err := stdErr.Write(string(b[0:n]))
		}
	}(l)

	go func(w io.Writer) error {
		b := make([]byte, 8192)
		for {
			n, err := w.Write(b)
			if err != nil {
				// these errors should be passed somewhere in channels...right now they are not handled anywhere.
				return errors.Wrap(err, join("", "Failed to write to the qemu socket ", socket, " from the buffer"))
			}
			_, err = in.Read(b[0:n]) // err is for short reads -- should probably retry in that case...
		}
	}(s)

	return nil
}

func (con *qemuConsole) pipeAttach(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	if con.machine().State != RUNNING {
		return nil, errors.New("kvm instance is not running")
	}

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
	rOut, wOut, err := os.Pipe() // stdout
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stdout")
	}
	fds = append(fds, rOut, wOut)
	con.fds = append(con.fds, rOut)
	rErr, wErr, err := os.Pipe() // stderr
	if err != nil {
		defer closeAll()
		return nil, errors.Wrap(err, "Failed os.Pipe for stderr")
	}
	fds = append(fds, rErr, wErr)
	con.fds = append(con.fds, rErr)

	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed, 3)
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(wIn, param.Stdin)
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
		closeChan <- err
		defer waitClosed.Done()
	}()
	waitClosed.Add(1)
	go func() {
		_, err := io.Copy(param.Stderr, rErr)
		closeChan <- err
		defer waitClosed.Done()
	}()
	go func() {
		waitClosed.Wait()
		defer close(closeChan)
	}()

	if err := con.unixSocketConsole(rIn, wOut, wErr); err != nil {
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
	}()

	return closeChan, nil
}

func (d *QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &qemuConsole{
		qemu: d,
	}
}

func (con *qemuConsole) Wait() error {
	return nil
}

func (con *qemuConsole) ForceClose() error {
	return nil
}
