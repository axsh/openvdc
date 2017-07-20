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

// TODO: add more types?
type CommandType int

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

type GuestAgentCommandArgs struct {
	Pid             int      `json:"pid,omitempty"`
	Path            string   `json:"path,omitempty"`
	Args            []string `json:"arg,omitempty"`
	EnvironmentVars []string `json:"env,omitempty"`
	CaptureOutput   bool     `json:"capture-output,omitempty"`
}

type GuestAgentRequest struct {
	Command   string                 `json:"execute"`
	Arguments *GuestAgentCommandArgs `json:"arguments"`
}

type GuestAgentResponse struct {
	Return *GuestAgentCommandResponse `json:"return"`
}

type GuestAgentCommandResponse struct {
	Pid      int    `json:"pid,omitempty"`
	Exited   bool   `json:"exited,omitempty"`
	Signal   int    `json:"signal,omitempty"`
	Exitcode int    `json:"exitcode,omitempty"`
	Stdout   string `json:"out-data,omitempty"`
	Stderr   string `json:"err-data,omitempty"`
}

func NewGuestAgentRequest(cmdType CommandType) *GuestAgentRequest {
	return &GuestAgentRequest{
		Command: guestCommand[cmdType],
	}
}

func NewGuestAgentExecStatusRequest(pid int) *GuestAgentRequest {
	return &GuestAgentRequest{
		Command: guestCommand[GuestExecStatus],
		Arguments: &GuestAgentCommandArgs{
			Pid: pid,
		},
	}
}

func NewGuestAgentExecRequest(cmd []string, output bool) *GuestAgentRequest {
	gaCmd := &GuestAgentRequest{
		Command: guestCommand[GuestExec],
		Arguments: &GuestAgentCommandArgs{
			Path:          cmd[0],
			CaptureOutput: output,
		},
	}

	if len(cmd) > 1 {
		gaCmd.Arguments.Args = make([]string, 0)
		for _, arg := range cmd[1:] {
			gaCmd.Arguments.Args = append(gaCmd.Arguments.Args, arg)
		}
	}

	return gaCmd
}

func (c *GuestAgentRequest) SendRequest(conn net.Conn) (*GuestAgentCommandResponse, error) {
	readBuf := bufio.NewReader(conn)
	errc := make(chan error)

	var err error

	sendRequest := func(cmd *GuestAgentRequest, resp *GuestAgentResponse) {
		var request []byte
		if request, err = json.Marshal(cmd); err != nil {
			errc <- errors.Wrap(err, "Failed to mashal json")
		}
		if _, err = fmt.Fprint(conn, strings.Join([]string{string(request), "\n"}, "")); err != nil {
			errc <- errors.Errorf("Failed to write to socket")
		}
		conn.SetReadDeadline(time.Now().Add(time.Second))
		var timeout time.Time
		go func() {
			timeout = <-time.After(time.Second * 10)
		}()

		for {
			line, _, _ := readBuf.ReadLine()
			if len(line) > 0 {
				if err = json.Unmarshal(line, &resp); err != nil {
					errc <- errors.Errorf("failed to unmarshal: %s", string(line))
					return
				}
				switch cmd.Command {
				case "guest-exec-status":
					if resp.Return.Exited {
						errc <- nil
						return
					}
					// guest-exec-status requires the command to have exited before the stdout/stderr/exitcode fields are returned
					if _, err = fmt.Fprint(conn, strings.Join([]string{string(request), "\n"}, "")); err != nil {
						errc <- errors.Errorf("Failed to write to socket")
					}
					continue
				default:
					errc <- nil
					return
				}
			}
			if !timeout.IsZero() {
				errc <- errors.Errorf("No response from agent, timed out: %s", string(request))
				return
			}
			time.Sleep(time.Second * 1)
		}
	}

	pidResp := &GuestAgentResponse{}
	go sendRequest(c, pidResp)
	if err = <-errc; err != nil {
		return nil, err
	}

	statusResp := &GuestAgentResponse{}
	go sendRequest(NewGuestAgentExecStatusRequest(pidResp.Return.Pid), statusResp)
	if err = <-errc; err != nil {
		return nil, err
	}

	return statusResp.Return, nil
}

func Base64toUTF8(str string) string {
	b, _ := base64.StdEncoding.DecodeString(str)
	return string(b)
}
