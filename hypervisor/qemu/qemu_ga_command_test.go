package qemu

import (
	"testing"
	"net"
	"bufio"

	"github.com/stretchr/testify/assert"
)

func TestNewGuestAgentExecCommand(t *testing.T) {
	assert := assert.New(t)
	req := NewGuestAgentExecCommand([]string{"cmd", "param1", "param2"}, true)

	assert.NotNil(req)
	assert.Equal(req.Command, "guest-exec")
	assert.Equal(req.Arguments.Path, "cmd")
	assert.Equal(req.Arguments.CaptureOutput, true)
	assert.Equal(len(req.Arguments.Args), 2)

	assert.Equal(req.Arguments.Args[0], "param1")
	assert.Equal(req.Arguments.Args[1], "param2")
}

func TestNewGuestAgentExecStatusCommand(t *testing.T) {
	assert := assert.New(t)
	req := NewGuestAgentExecStatusCommand(1)
	assert.NotNil(req)
	assert.Equal(req.Command, "guest-exec-status")
	assert.Equal(req.Arguments.Pid, 1)
}

func TestNewGuestAgentCommand(t *testing.T) {
	assert := assert.New(t)
	req := NewGuestAgentCommand(GuestPing)
	assert.NotNil(req)
	assert.Equal(req.Command, "guest-ping")
}

}
