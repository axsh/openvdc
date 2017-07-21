package qemu

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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

func TestSendRequest(t *testing.T) {
	assert := assert.New(t)
	resp := make(chan []byte)

	execResponse := `{"return":{"pid":100}}`
	execStatsuResponse := `{"return":{"exitcode":0,"exited":true,"out-data":"stdout"}}`

	socketPath := "/tmp/tmpsocket"
	listener, _ := net.Listen("unix", socketPath)
	connection, _ := net.Dial("unix", socketPath)

	var listenerConnection net.Conn
	var err error
	go func() {
		for {
			if listenerConnection, err = listener.Accept(); err != nil {
				fmt.Println(err)
				continue
			}
			defer listenerConnection.Close()
			go func(c net.Conn) {
				buf := bufio.NewReader(c)
				for {
					if line, err := buf.ReadBytes('\n'); err == nil {
						if strings.Contains(string(line), "guest-exec-status") {
							fmt.Fprintf(c, execStatsuResponse)
							resp <- line
						}
						if strings.Contains(string(line), "guest-exec") {
							fmt.Fprintf(c, execResponse)
							resp <- line
						}
					}
				}
			}(listenerConnection)
		}
	}()

	var resp1 GuestAgentResponse

	NewGuestAgentExecRequest([]string{"cmd", "arg1", "arg2"}, true).SendRequest(connection, &resp1)
	assert.Equal(resp1.Return.(*ExecResponse).Pid, 100)

	execRequest := `{"execute":"guest-exec","arguments":{"path":"cmd","arg":["arg1","arg2"],"capture-output":true}}`
	assert.Equal(strings.Join([]string{execRequest, "\n"}, ""), string(<-resp))

	var resp2 GuestAgentResponse

	NewGuestAgentExecStatusRequest(100).SendRequest(connection, &resp2)
	assert.Equal(resp2.Return.(*ExecStatusResponse).Exitcode, 0)
	assert.Equal(resp2.Return.(*ExecStatusResponse).Exited, true)
	assert.Equal(resp2.Return.(*ExecStatusResponse).Stdout, "stdout")

	execStatusRequest := `{"execute":"guest-exec-status","arguments":{"pid":100}}`
	assert.Equal(strings.Join([]string{execStatusRequest, "\n"}, ""), string(<-resp))

	listener.Close()
	connection.Close()
}
