// +build acceptance

// +build acceptance

package tests

import (
	"strings"
	"testing"
	"os/exec"
	"time"

	"github.com/tidwall/gjson"
)

func TestCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	waitUntil(t, 5 * time.Minute, func()error {
		cmd := exec.Command("openvdc", "show", instance_id)
		buf, err := cmd.CombinedOutput()
		if err !=nil {
			return err
		}
		result := gjson.GetBytes(buf, "instance.lastState.state")
		if result.String() == "RUNNING" {
			return nil
		}
		time.Sleep(5 * time.Second)
		return WaitContinue
	})

	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
}
