package qemu

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type CommandType int

// TODO: add more types?
const (
	GuestExec CommandType = iota // 0
	GuestExecStatus
	GuestPing
)

var guestCommand = map[CommandType]string{
	GuestExec:       "guest-exec",
	GuestExecStatus: "guest-exec-status",
	GuestPing:       "guest-ping",
}

var guestCommandType = map[string]CommandType{
	"guest-exec":        GuestExec,
	"guest-exec-status": GuestExecStatus,
	"guest-ping":        GuestPing,
}

type GuestAgentRequest struct {
	Command   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type GuestAgentResponse struct {
	Return interface{} `json:"return"`
}

type ExecStatusResponse struct {
	Exited   bool      `json:"exited,omitempty"`
	Signal   int       `json:"signal,omitempty"`
	Exitcode int       `json:"exitcode,omitempty"`
	Stdout   string    `json:"out-data,omitempty"`
	Stderr   string    `json:"err-data,omitempty"`
	TruncStdOut string `json:"out-truncated,omitempty"`
	TruncStdErr string `json:"err-truncated,omitempty"`
}

type ExecStatusRequest struct {
	Pid int `json:"pid,omitempty"`
}

type ExecResponse struct {
	Pid int `json:"pid,omitempty"`
}

type ExecRequest struct {
	Path            string   `json:"path,omitempty"`
	Args            []string `json:"arg,omitempty"`
	EnvironmentVars []string `json:"env,omitempty"`
	CaptureOutput   bool     `json:"capture-output,omitempty"`
}

func NewGuestAgentRequest(cmdType CommandType) *GuestAgentRequest {
	return &GuestAgentRequest{
		Command: guestCommand[cmdType],
	}
}

func NewGuestAgentExecStatusRequest(pid int) *GuestAgentRequest {
	return &GuestAgentRequest{
		Command: guestCommand[GuestExecStatus],
		Arguments: &ExecStatusRequest{
			Pid: pid,
		},
	}
}

func NewGuestAgentExecRequest(cmd []string, output bool) *GuestAgentRequest {
	execReq := &ExecRequest{
		Path:          cmd[0],
		CaptureOutput: output,
	}
	if len(cmd) > 1 {
		execReq.Args = make([]string, 0)
		for _, arg := range cmd[1:] {
			execReq.Args = append(execReq.Args, arg)
		}
	}

	return &GuestAgentRequest{
		Command: guestCommand[GuestExec],
		Arguments: execReq,
	}
}

func (c *GuestAgentRequest) SendRequest(conn net.Conn, response *GuestAgentResponse) error {
	readBuf := bufio.NewReader(conn)
	errc := make(chan error)

	var err error

	go func() {
		var request []byte
		var timeout time.Time

		unmarshal := func(data []byte, out *GuestAgentResponse) {
			if err = json.Unmarshal(data, &out); err != nil {
				errc <- errors.Errorf("failed to unmarshal: %s", string(data))
			}
		}

		go func() {
			timeout = <-time.After(time.Second * 10)
		}()
		if request, err = json.Marshal(c); err != nil {
			errc <- errors.Wrap(err, "Failed to mashal json")
		}
		if _, err = fmt.Fprint(conn, strings.Join([]string{string(request), "\n"}, "")); err != nil {
			errc <- errors.Errorf("Failed to write to socket")
		}
		conn.SetReadDeadline(time.Now().Add(time.Second))
		for {
			line, _, _ := readBuf.ReadLine()
			if len(line) <= 0 {
				if !timeout.IsZero() {
					errc <- errors.Errorf("No response from agent, timed out: %s", string(request))
					return
				}
				time.Sleep(time.Second * 1)
			}

			switch guestCommandType[c.Command] {
			case GuestExecStatus:
				response.Return = &ExecStatusResponse{}
				unmarshal(line, response)
				// TODO: handle truncated outputs
				if response.Return.(*ExecStatusResponse).Exited {
					errc <- nil
					return
				}
				// guest-exec-status requires the command to have exited before the stdout/stderr/exitcode fields are returned
				if _, err = fmt.Fprint(conn, strings.Join([]string{string(request), "\n"}, "")); err != nil {
					errc <- errors.Errorf("Failed to write to socket")
					return
				}
			case GuestExec:
				response.Return = &ExecResponse{}
				unmarshal(line, response)
				errc <- nil
				return
			}
		}
	}()
	if err = <-errc; err != nil {
		return err
	}
	return nil
}

func Base64toUTF8(str string) string {
	b, _ := base64.StdEncoding.DecodeString(str)
	return string(b)
}
