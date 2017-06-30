package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"net"
	"github.com/pkg/errors"
	"bufio"
	"time"
	"strings"
)

type State int

const (
	STOPPED State = iota // 0
	STARTING
	RUNNING
	STOPPING
	REBOOTING
	SHUTTINGDOWN
	TERMINATING
	FAILED
)

var stateValues = map[State]string{
	STOPPED : "STOPPED",
	STARTING : "STARTING",
	RUNNING : "RUNNING",
	STOPPING : "STOPPING",
	REBOOTING : "REBOOTING",
	SHUTTINGDOWN : "SHUTTINGDOWN",
	FAILED : "FAILED",
}

var MachineState = map[string]State{
	"STOPPED" : STOPPED,
	"STARTING" : STARTING,
	"RUNNING" : RUNNING,
	"STOPPING" : STOPPING,
	"REBOOTING" : REBOOTING,
	"SHUTTINGDOWN" : SHUTTINGDOWN,
	"TERMINATING" : TERMINATING,
	"FAILED" : FAILED,
}

type Machine struct {
	State   State
	Cores   int
	Memory  uint64
	Name    string
	Display string
	Vnc     string
	Monitor string
	Serial  string
	Pidfile string
	Nics    []NetDev
	Drives  map[string]Drive
	Process *os.Process
	Kvm     bool
}

type NetDev struct {
	IfName       string
	Index        string
	Ipv4Addr     string
	MacAddr      string
	Bridge       string
	BridgeHelper string
	Type         string
}

func startStateEvaluation(timeout time.Duration, evaluationFunction func() bool) <-chan bool {
	passed := make(chan bool, 1)
	timeoutc := make(chan bool, 1)

	go func() {
		time.Sleep(timeout)
		timeoutc <-true
	}()

	go func() {
		for {
			select {
			case <-timeoutc:
				passed <-false
				return
			default:
				if evaluationFunction() {
					passed <-true
					return
				}
			}
		}
	}()
	return passed
}


func (m *Machine) HavePrompt() bool {
	c, err := net.Dial("unix", m.Serial)
	buf := bufio.NewReader(c)
	defer c.Close()

	if err != nil {
		return false
		// return errors.Errorf("Failed to connect to serial socket: %s:", d.machine.Serial)
	}
	if err := c.SetReadDeadline(time.Now().Add(5*time.Second)); err != nil {
		return false
	}

	b, _ := buf.ReadBytes('\n')
	fmt.Println(string(b))
	return (strings.Contains(string(b), m.Name))
}

func (m *Machine) ScheduleState(nextState State, timeout time.Duration, callback func() bool) error {
	// if m.State == nextState {
	// 	return errors.Errorf("Already in state %s", stateValues[nextState])
	// }
	passed := <-startStateEvaluation(timeout, callback)
	if !passed {
		return errors.Errorf("Timed out scheduling state %s", stateValues[nextState])
	}

	fmt.Println("Setting state: " +stateValues[nextState])
	m.State = nextState
	return nil
}

func NewMachine(cores int, mem uint64) *Machine {
	return &Machine{
		Cores: cores,
		Memory: mem,
		Drives: make(map[string]Drive),
		Display: "none",
	}
}

func (m *Machine) AddNICs(nics []NetDev) {
	for _, nic := range nics {
		m.Nics = append(m.Nics, nic)
	}
}

func (m *Machine) Start(startCmd string) error {
	qemuCmd := fmt.Sprintf("%s", startCmd)
	cmdLine := &cmdLine{args: make([]string, 0)}

	cmd := exec.Command(qemuCmd, cmdLine.QemuBootCmd(m)...)
	fmt.Println(cmd.Args)
	if  err := cmd.Run(); err != nil {
		return errors.Errorf("Failed to execute cmd: %s", cmd.Args)
	}

	m.Process = cmd.Process
	// TODO: add some error handling

	return m.ScheduleState(STARTING, (30*time.Minute), func() bool {
		err := m.MonitorCommand("info name")
		return (err != nil)
	})
}

func (m *Machine) MonitorCommand(cmd string) error {
	c, err := net.Dial("unix", m.Monitor)
	if err != nil {
		return errors.Errorf("Failed to connect to monitor socket %s:", m.Monitor)
	}
	defer c.Close()

	fmt.Fprintf(c, "%s\n", cmd)
	return nil
}

func (m *Machine) Stop() error {
	if err := m.MonitorCommand("quit"); err != nil {
		return err
	}

	os.Remove(m.Monitor)
	os.Remove(m.Serial)

	return nil
}

func (m *Machine) Reboot() error {
	m.MonitorCommand("system_reset")
	return m.ScheduleState(RUNNING, (30*time.Second), func() bool {
		return true
	})
}
