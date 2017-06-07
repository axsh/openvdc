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

	cmd := exec.Command(qemuCmd, cmdLine.qemuBootCmd(m, true)...)
	return nil
}

