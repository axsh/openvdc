package util

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
)

type SerialConnection struct {
	SerialConn net.Conn
}

type serialConsoleParam struct {
	remoteConsole *hypervisor.ConsoleParam
	waitClosed    *sync.WaitGroup
	errc          chan error
	closeChan     chan bool
}

func (sc *SerialConnection) stdinToConn(param *serialConsoleParam, finished <-chan struct{}) {
	param.waitClosed.Add(1)
	defer func() {
		param.closeChan <-true
		param.waitClosed.Done()
	}()

	b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
	for {
		select {
		case <-finished:
			return
		case err := <-param.errc:
			param.errc <- err
			break
		default:
			n, err := param.remoteConsole.Stdin.Read(b)
			if err != nil {
				if err == io.EOF {
					log.Info("\nConsole exited by EOF\n\n")
					param.errc <- nil
				} else {
					param.errc <- errors.Wrap(err, "\nFailed to read from the from the console input buffer\n\n")
				}
			}

			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.Info("\nConsole exited by ctrl-q\n\n")
				param.errc <- nil
			}
			_, err = sc.SerialConn.Write(b[0:n])
			if err != nil {
				param.errc <- errors.Wrap(err, "\nFailed to write to the connection from the buffer\n\n")
			}
		}
	}
}

func (sc *SerialConnection) connToStdout(param *serialConsoleParam, finished <-chan struct{}) {
	param.waitClosed.Add(1)
	defer func() {
		param.closeChan <-true
		param.waitClosed.Done()
	}()

	b := make([]byte, 8192)
	for {
		sc.SerialConn.SetDeadline(time.Now().Add(time.Second))
		n, err := sc.SerialConn.Read(b)

		select {
		case <-finished:
			return
		case err := <-param.errc:
			if _, e := param.remoteConsole.Stdout.Write([]byte{0x0A}); e != nil {
				param.errc <- errors.Wrap(e, "\nFailed to write the linefeed character on exit\n\n")
			} else {
				param.errc <- err
			}
			break
		default:
			if err != nil && !err.(net.Error).Timeout() {
				param.errc <- errors.Wrap(err, "\nFailed to read the connection buffer from socket ")
			}
			// exit on ctrl + q
			if bytes.Contains(b[0:n], []byte{0x11}) {
				param.errc <- nil
			}

			if synchable, ok := param.remoteConsole.Stdout.(*os.File); ok {
				if err := synchable.Sync(); err != nil {
					log.Warn("Failed %v to flush %s", param.remoteConsole.Stdout, err)
				}
			}
			_, err = param.remoteConsole.Stdout.Write(b[0:n])
			if err != nil {
				errc <- errors.Wrap(err, "\nFailed to write to the console stdout buffer\n\n")
			}
		}
	}
}

func (sc *SerialConnection) AttachSerialConsole(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup, errc chan error) error {
	finished := make(chan struct{})
	serialConsoleParam := serialConsoleParam{
		remoteConsole: param,
		waitClosed:    waitClosed,
		errc:          errc,
		closeChan:     make(chan bool),
	}

	go sc.stdinToConn(&serialConsoleParam, finished)
	go sc.connToStdout(&serialConsoleParam, finished)
	go func() {
		<-serialConsoleParam.closeChan
		close(finished)
	}()
	return nil
}
