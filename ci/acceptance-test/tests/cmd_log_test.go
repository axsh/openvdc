// +build acceptance

package tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"
)

func TestOpenVDCLog(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	defer RunCmd("openvdc", "destroy", instance_id)

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

	RunCmdAndReportFail(t, "openvdc", "log", instance_id)
}
