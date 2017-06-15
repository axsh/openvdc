// build +linux

package qemu

import (
	"testing"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestQemuBootCmd(t *testing.T) {
	assert := assert.New(t)
	cmdLine := &cmdLine{args: make([]string, 0)}
	cmd := cmdLine.QemuBootCmd(&Machine{
		Cores: 1,
		Memory: 512,
		Name: "machine",
		Monitor: "monitor",
		Serial: "serial",
		Nics: []NetDev{
			NetDev{
				IfName: "if",
				MacAddr: "mac",
			},
		},
		Drives: []Drive{
			Drive{
				Image: &Image{
					Format: "raw",
					Path: "drive",
				},
				If: "disk",
			},
		},
		Kvm: true,
		Display: "none",
	})

	assert.Equal(strings.Join(cmd, " "), "-smp 1 -m 512 -enable-kvm -serial unix:serial,server,nowait -monitor unix:monitor,server,nowait -drive file=drive,format=raw,if=disk -netdev tap,ifname=if,id=if -device virtio-net-pci,netdev=if,mac=mac -display none")

}

func TestappendArgs(t *testing.T) {
	assert := assert.New(t)
	cmdLine := &cmdLine{args: make([]string, 0)}
	cmdLine.appendArgs("arg0")
	cmdLine.appendArgs("arg1", "arg2")

	assert.Equal(cmdLine.args[0], "arg0")
	assert.Equal(cmdLine.args[1], "arg1")
	assert.Equal(cmdLine.args[2], "arg2")
}

func TestQemuImgCmd(t *testing.T) {
	assert := assert.New(t)

	cmdLine1 := &cmdLine{args: make([]string, 0)}
	cmd1 := cmdLine1.QemuImgCmd(&Image{
		Format: "raw",
		baseImg: "baseImage",
		Path: "path",
	})

	cmdLine2 := &cmdLine{args: make([]string, 0)}
	cmd2 := cmdLine2.QemuImgCmd(&Image{
		Format: "raw",
		Size: 1000,
		Path: "path",
	})

	assert.Equal(strings.Join(cmd1, " "), "create -f raw path -b baseImage")
	assert.Equal(strings.Join(cmd2, " "), "create -f raw path 1000K")
}
