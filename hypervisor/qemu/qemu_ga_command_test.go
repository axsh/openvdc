package qemu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQEMUExecCommand(t *testing.T) {
	assert := assert.New(t)
	qcmd := NewQEMUExecCommand([]string{"cmd", "param1", "param2"}, true)

	assert.NotNil(qcmd)
	assert.Equal(qcmd.Command, "guest-exec")
	assert.Equal(qcmd.Arguments.Path, "cmd")
	assert.Equal(qcmd.Arguments.CaptureOutput, true)
	assert.Equal(len(qcmd.Arguments.Args), 2)

	assert.Equal(qcmd.Arguments.Args[0], "param1")
	assert.Equal(qcmd.Arguments.Args[1], "param2")
}

func TestNewQEMUExecStatusCommand(t *testing.T) {
	assert := assert.New(t)
	qcmd := NewQEMUExecStatusCommand(1)
	assert.NotNil(qcmd)
	assert.Equal(qcmd.Command, "guest-exec-status")
	assert.Equal(qcmd.Arguments.Pid, 1)
}

func TestNewQEMUCommand(t *testing.T) {
	assert := assert.New(t)
	qcmd := NewQEMUCommand(GuestPing)
	assert.NotNil(qcmd)
	assert.Equal(qcmd.Command, "guest-ping")
}
