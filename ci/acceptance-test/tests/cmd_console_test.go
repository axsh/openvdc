// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "RUNNING")
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "TERMINATED")
}
