package qemu

import (
	"fmt"
	"strconv"
	"os/exec"
)

type Machine struct {
	Cores   int
	Memory  uint64
	Name    string
	Display string
	Vnc     string
	Monitor string
	Serial  string
	Nics    []NetDev
	Drives  []Drive
	Pid     int
}

type NetDev struct {
	IfName       string
	Index        string
	MacAddr      string
	Bridge       string
	BridgeHelper string
}

func NewMachine(cores int, mem uint64) *Machine {
	return &Machine{
		Cores: cores,
		Memory: mem,
		Drives: make([]Drive, 0),
	}
}

func (m *Machine) AddDrive(d Drive) {
	m.Drives = append(m.Drives, d)
}

func (m *Machine) AddNICs(nics []NetDev) {
	for _, nic := range nics {
		m.Nics = append(m.Nics, nic)
	}
}

func (m *Machine) Start(startCmd string) error {
	qemuCmd := fmt.Sprintf("%s", startCmd)
	cmdLine := &cmdLine{args: make([]string, 0)}

	cmd := exec.Command(qemuCmd, cmdLine.buildQemuCmd(m, true)...)
	fmt.Printf("%s", cmd.Args)
	if err := cmd.Start() ; err != nil {
		return errors.Errorf("Failed to execute cmd: %s", cmd.Args)
	}

	m.Pid = cmd.Process.Pid

	// TODO: add some error handling
	return nil
}
