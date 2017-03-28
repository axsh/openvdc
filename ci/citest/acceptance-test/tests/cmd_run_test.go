// +build acceptance

package tests

import (
	"strings"
	"testing"
)

func TestCmdRun_NullTemplate(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null")
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "RUNNING")

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	RunCmdAndReportFail(t, "openvdc", "wait", instance_id, "TERMINATED")
}

func TestCmdRun_NoneTemplate(t *testing.T) {
	RunCmdAndExpectFail(t, "openvdc", "run", "none")
}
