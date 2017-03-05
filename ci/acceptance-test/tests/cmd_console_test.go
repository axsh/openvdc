// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("echo 'ls' | openvdc console %s", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("echo 'false' | openvdc console %s", instance_id))
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
