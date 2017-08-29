package esxi

import (
	"bytes"
	"github.com/axsh/openvdc/hypervisor"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type esxiConsole struct {
	esxi       *EsxiHypervisorDriver
	telnetConn net.Conn
	conChan    chan error
}

func (d *EsxiHypervisorDriver) InstanceConsole() hypervisor.Console {
	return &esxiConsole{
		esxi: d,
	}
}

func (c *esxiConsole) pipeAttach(param *hypervisor.ConsoleParam, args ...string) (<-chan hypervisor.Closed, error) {
	//TODO: check if machine is running {
	// return nil, errors.New("esxi instance is not in a running state")
	//}
	var err error
	waitClosed := new(sync.WaitGroup)
	closeChan := make(chan hypervisor.Closed)
	c.conChan = make(chan error)

	if len(args) == 0 {
		c.telnetConn, err = net.Dial("tcp", strings.Join([]string{settings.EsxiIp, strconv.Itoa(c.esxi.machine.SerialConsolePort)}, ":"))
		if err != nil {
			return nil, errors.Errorf("Unable to connect to %s on port %d", settings.EsxiIp, c.esxi.machine.SerialConsolePort)
		}
		go func() {
			waitClosed.Wait()
		}()
		c.attachShell(param, waitClosed)
	} else {
		c.execCommand(param, waitClosed, args...)
	}

	defer close(closeChan)
	return closeChan, nil
}

// TODO: this is essentially the same function as is used for the qemu serial console and should
// probably be refactored
func (c *esxiConsole) attachShell(param *hypervisor.ConsoleParam, waitClosed *sync.WaitGroup) error {
	waitClosed.Add(1)
	go func() {
		defer waitClosed.Done()
		b := make([]byte, 8192) // 8 kB is the default page size for most modern file systems
		for {
			select {
			case err := <-c.conChan:
				c.conChan <- err
				break
			default:
				n, err := param.Stdin.Read(b)
				if err != nil {
					c.conChan <- errors.Wrap(err, "\nFailed to read from the from the console input buffer\n\n")
				}
				// fmt.Println(join("", string(b[0:n]), " - from stdin"))

				if bytes.Contains(b[0:n], []byte{0x11}) {
					log.Info("Received exit from stdin")
					c.conChan <- errors.Wrap(err, "\nConsole exited by ctrl-q\n\n")
				}
				_, err = c.telnetConn.Write(b[0:n])
				if err != nil {
					c.conChan <- errors.Wrap(err, "\nFailed to write to telnet connection from the buffer\n\n")
				}
			}
		}
	}()

	waitClosed.Add(1)
	go func() {
		defer waitClosed.Done()
		b := make([]byte, 8192)
		for {
			c.telnetConn.SetDeadline(time.Now().Add(time.Second))
			n, err := c.telnetConn.Read(b)
			select {
			case err := <-c.conChan:
				if _, e := param.Stdout.Write([]byte{0x0A}); e != nil {
					c.conChan <- errors.Wrap(e, "\nFailed to write the linefeed character on exit\n\n")
				} else {
					c.conChan <- err
				}
				break
			default:
				if err != nil && !err.(net.Error).Timeout() {
					c.conChan <- errors.Wrap(err, "\nFailed to read from telnet connection\n")
				}
				// exit on ctrl + q
				if bytes.Contains(b[0:n], []byte{0x11}) {
					c.conChan <- nil
				}
				_, err = param.Stdout.Write(b[0:n])
				if err != nil {
					c.conChan <- errors.Wrap(err, "\nFailed to write to the console stdout buffer\n\n")
				}
			}
		}
	}()

	return nil
}

func (c *esxiConsole) execCommand(param *hypervisor.ConsoleParam, waitDone *sync.WaitGroup, args ...string) {
	c.conChan <- nil
}

func (c *esxiConsole) Attach(param *hypervisor.ConsoleParam) (<-chan hypervisor.Closed, error) {
	return c.pipeAttach(param)
}

func (c *esxiConsole) Exec(param *hypervisor.ConsoleParam, args []string) (<-chan hypervisor.Closed, error) {
	return c.pipeAttach(param, args...)
}

func (c *esxiConsole) Wait() error {
	defer c.telnetConn.Close()
	return <-c.conChan
}

func (c *esxiConsole) ForceClose() error {
	return nil
}

type consoleWaitError struct {
}

func (e *consoleWaitError) ExitCode() int {
	return 0
}
