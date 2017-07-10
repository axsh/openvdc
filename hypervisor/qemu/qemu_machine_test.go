// build +linux

package qemu

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMachine(t *testing.T) {
	assert := assert.New(t)
	machine := NewMachine(1, 512)
	assert.NotNil(machine)
	assert.Equal(1, machine.Cores)
	assert.Equal(uint64(512), machine.Memory)
	assert.NotNil(machine.Drives)
}

func TestAddNICs(t *testing.T) {
	assert := assert.New(t)
	machine := NewMachine(1, 512)
	machine.AddNICs([]Nic{Nic{IfName: "if0"}, Nic{IfName: "if1"}})

	assert.Equal(len(machine.Nics), 2)
	for idx, nic := range machine.Nics {
		assert.Equal(fmt.Sprintf("if%d", idx), nic.IfName)
	}
}
