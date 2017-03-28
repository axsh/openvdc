// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
)

func TestCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "RUNNING")
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("echo 'ls' | openvdc console %s", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("echo 'false' | openvdc console %s", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ls", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- ls", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- false", instance_id))
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "TERMINATED")
}
