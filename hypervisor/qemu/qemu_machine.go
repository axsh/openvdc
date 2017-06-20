package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"net"
	"github.com/pkg/errors"
)

type State int

const (
	STOPPED State = iota
	STARTING
	RUNNING
	STOPPING
	REBOOTING
	SHUTTINGDOWN
	TERMINATING
	FAILED
)

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
	Nics    []NetDev
	Drives  map[string]Drive
	Process *os.Process
	Kvm     bool
}

type NetDev struct {
	IfName       string
	Index        string
	MacAddr      string
	Bridge       string
	BridgeHelper string
}

func (m *Machine) scheduleState(nextState State, timeout int) error {
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
	return nil
}

func (m *Machine) MonitorCommand(cmd string) error {
	c, err := net.Dial("unix", fmt.Sprintf("%s", m.Monitor))
	fmt.Println(m.Monitor)
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
	return nil
}