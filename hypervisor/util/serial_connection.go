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
	var err error
	param.waitClosed.Add(1)
	defer func() {
		param.closeChan <-true
		param.errc <-err
		param.waitClosed.Done()
	}()

	b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
	for {
		select {
		case <-finished:
			return
		default:
			n, err := param.remoteConsole.Stdin.Read(b)
			if err != nil {
				if err == io.EOF {
					log.Info("Release serial console, received EOF")
				} else {
					log.WithError(err).Error("Failed to read from the from the console input buffer")
				}
				return
			}

			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.Info("Release serial console, received escape code (ctrl-q)")
				return
			}

			_, err = sc.SerialConn.Write(b[0:n])
			if err != nil {
				log.WithError(err).Error("Failed to write to the connection from the buffer")
				return
			}
		}
	}
}

func (sc *SerialConnection) connToStdout(param *serialConsoleParam, finished <-chan struct{}) {
	var err error
	param.waitClosed.Add(1)
	defer func() {
		if _, err = param.remoteConsole.Stdout.Write([]byte{0x0A}); err != nil {
			if err == io.EOF {
				err = nil
			} else {
				log.WithError(err).Error("Failed to write the linefeed character on exit")
			}
		}

		param.closeChan <-true
		param.errc <-err
		param.waitClosed.Done()
	}()

	b := make([]byte, 8192)
	for {
		sc.SerialConn.SetDeadline(time.Now().Add(time.Second))
		n, err := sc.SerialConn.Read(b)

		select {
		case <-finished:
			return
		default:
			if err == io.EOF {
				_, err = param.remoteConsole.Stdout.Write([]byte("\nConnection lost\n"))
				if err != nil && err != io.EOF {
					log.WithError(err).Error("Failed to write the linefeed character on exit")
					return
				}

				log.Info("Connection released, received EOF")
				return
			}
			if err != nil && !err.(net.Error).Timeout() {
				log.WithError(err).Error("Failed to read the connection buffer from socket")
				return
			}
			// exit on ctrl + q
			if bytes.Contains(b[0:n], []byte{0x11}) {
				return
			}
			if synchable, ok := param.remoteConsole.Stdout.(*os.File); ok {
				if err := synchable.Sync(); err != nil {
					log.Warn("Failed %v to flush %s", param.remoteConsole.Stdout, err)
				}
			}
			_, err = param.remoteConsole.Stdout.Write(b[0:n])
			if err != nil {
				log.WithError(err).Error("Failed to write to the console stdout buffer")
				return
			}
		}
	}
}

func (sc *SerialConnection) AttachSerialConsole(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup, errc chan error) error {
	if sc.SerialConn == nil {
		return errors.Errorf("Failed to attach console, connection not available")
	}

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
