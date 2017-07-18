package qemu

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"bufio"
	"time"
	"strings"

	"github.com/pkg/errors"
)

type QEMUCommandArgs struct {
	Pid             int      `json:"pid,omitempty"`
	Path            string   `json:"path,omitempty"`
	Args            []string `json:"arg,omitempty"`
	EnvironmentVars []string `json:"env,omitempty"` 
	CaptureOutput   bool     `json:"capture-output,omitempty"`
}

type QEMUCommand struct {
	Command   string           `json:"execute"`
	Arguments *QEMUCommandArgs `json:"arguments"`
}

type QEMUResponse struct {
	Return *QEMUCommandResponse `json:"return"`
}

type QEMUCommandResponse struct {
	Pid      int    `json:"pid,omitempty"`
	Exited   bool   `json:"exited,omitempty"`
	Signal   int    `json:"signal,omitempty"`
	Exitcode int    `json:"exitcode,omitempty"`
	Stdout   string `json:"out-data,omitempty"`
	Stderr   string `json:"err-data,omitempty"`
}

func NewQEMUCommand(cmd []string, output bool) *QEMUCommand {
	gaCmd := &QEMUCommand{
		Command: "guest-exec",
		Arguments: &QEMUCommandArgs{
			Path: cmd[0],
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

func (c *QEMUCommand) convertToJson() (string, error) {
	query, err := json.Marshal(c)
	if err != nil {
		return "",err
	}

	return string(query), nil
}


func (c *QEMUCommand) SendCommand(conn net.Conn) (*QEMUCommandResponse, error) {
	readBuf := bufio.NewReader(conn)
	errc := make(chan error)
	sendQuery := func(cmd *QEMUCommand, resp *QEMUResponse) {
		query, err := cmd.convertToJson()
		if err != nil {
			errc <- errors.Wrap(err, "Failed to mashal json")
		}
		if _, err := fmt.Fprint(conn, strings.Join([]string{query, "\n"}, "")); err != nil {
			errc <- errors.Errorf("Failed to write to socket")
		}
		conn.SetReadDeadline(time.Now().Add(time.Second))
		var timeout time.Time
		go func() {
			timeout = <- time.After(time.Second*10)
		}()

		for {
			line, _, _ := readBuf.ReadLine()
			if len(line) > 0 {
				if err := json.Unmarshal(line, &resp); err != nil {
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
					if _, err := fmt.Fprint(conn, strings.Join([]string{query, "\n"}, "")); err != nil {
						errc <- errors.Errorf("Failed to write to socket")
					}
					continue
				default:
					errc <- nil
					return
				}
			}
			if (!timeout.IsZero()) {
				errc <- errors.Errorf("No response from agent, timed out: %s", query)
				return
			}
			time.Sleep(time.Second*1)
		}
	}

	pidResp := &QEMUResponse{}
	go sendQuery(c, pidResp)
	if err := <-errc; err != nil {
		return nil, err
	}

	statusResp := &QEMUResponse{}
	go sendQuery(&QEMUCommand{Command: "guest-exec-status", Arguments: &QEMUCommandArgs{Pid: pidResp.Return.Pid}}, statusResp)
	if err := <-errc; err != nil {
		return nil, err
	}

	return statusResp.Return, nil
}

func Base64toUTF8(str string) string {
	b, _ := base64.StdEncoding.DecodeString(str)
	return string(b)
}
