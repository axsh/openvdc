package qemu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGuestAgentExecRequest(t *testing.T) {
	assert := assert.New(t)
	req := NewGuestAgentExecRequest([]string{"cmd", "param1", "param2"}, true)

	assert.NotNil(req)
	assert.IsType((*ExecRequest)(nil), req.Arguments)

	arg := req.Arguments.(*ExecRequest)
	assert.Equal(req.Command, "guest-exec")
	assert.Equal(arg.Path, "cmd")
	assert.Equal(arg.CaptureOutput, true)
	assert.Equal(len(arg.Args), 2)

	assert.Equal(arg.Args[0], "param1")
	assert.Equal(arg.Args[1], "param2")
}

func TestNewGuestAgentExecStatusRequest(t *testing.T) {
	assert := assert.New(t)
	req := NewGuestAgentExecStatusRequest(1)

	assert.NotNil(req)
	assert.IsType((*ExecStatusRequest)(nil), req.Arguments)

	arg := req.Arguments.(*ExecStatusRequest)
	assert.Equal(req.Command, "guest-exec-status")
	assert.Equal(arg.Pid, 1)
}

func TestNewGuestAgentRequest(t *testing.T) {
	assert := assert.New(t)

	req := NewGuestAgentRequest(GuestPing)
	assert.NotNil(req)
	assert.Equal(req.Command, "guest-ping")
}
