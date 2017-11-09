package util

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
)

type SerialConnection struct {
	SerialConn net.Conn
	err        error
}

type serialConsoleParam struct {
	remoteConsole *hypervisor.ConsoleParam
	waitClosed    *sync.WaitGroup
	errc          chan error
	closeChan     chan bool
}

func (p *serialConsoleParam) flushConsole() {
	if synchable, ok := p.remoteConsole.Stdout.(*os.File); ok {
		if err := synchable.Sync(); err != nil {
			log.Warn("Failed %v to flush %s", p.remoteConsole.Stdout, err)
		}
	}
}

func (p *serialConsoleParam) resetConsole() {
	p.flushConsole()
	_, err := p.remoteConsole.Stdout.Write([]byte{0x0A, 0x0D})
	if err != nil && err != io.EOF {
		log.WithError(err).Warn("Failed to write the linefeed/character return")
	}
}

func (sc *SerialConnection) stdinToConn(param *serialConsoleParam, finished <-chan struct{}) {
	param.waitClosed.Add(1)
	defer func() {
		param.closeChan <-true
		if sc.err != nil {
			param.errc <-sc
		} else {
			param.errc <-nil
		}
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
					param.resetConsole()
				} else {
					log.WithError(err).Error("Failed to read from the from the console input buffer")
				}
				sc.err = err
				return
			}

			if bytes.Contains(b[0:n], []byte{0x11}) {
				log.Info("Release serial console, received escape code (ctrl-q)")
				param.resetConsole()
				return
			}

			if _, err := sc.SerialConn.Write(b[0:n]); err != nil {
				log.WithError(err).Error("Failed to write to the connection from the buffer")
				sc.err = err
				return
			}
		}
	}
}

func (sc *SerialConnection) connToStdout(param *serialConsoleParam, finished <-chan struct{}) {
	param.waitClosed.Add(1)
	defer func() {
		param.closeChan <-true
		if sc.err != nil {
			param.errc <-sc
		} else {
			param.errc <-nil
		}
		param.waitClosed.Done()
	}()

	b := make([]byte, 8192)
	for {
		sc.SerialConn.SetDeadline(time.Now().Add(time.Second))
		n, err := sc.SerialConn.Read(b)
		if err == io.EOF {
			sc.err = errors.New("connection lost")
			param.resetConsole()
			return
		}

		select {
		case <-finished:
			return
		default:
			if err != nil && !err.(net.Error).Timeout() {
				log.WithError(err).Error("Failed to read the connection buffer from socket")
				sc.err = err
				return
			}

			param.flushConsole()
			_, err = param.remoteConsole.Stdout.Write(b[0:n])
			if err != nil{
				log.WithError(err).Error("Failed to write to the console stdout buffer")
				sc.err = err
				return
			}
		}
	}
}

func (sc *SerialConnection) Error() string {
	return fmt.Sprintf("Serial console failed: %v", sc.err)
}

func (sc *SerialConnection) ExitCode() int {
	if sc.err != nil && sc.err != io.EOF {
		return 1
	}
	return 0
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
