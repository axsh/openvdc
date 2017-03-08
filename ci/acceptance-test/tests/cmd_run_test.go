// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCmdRun_NullTemplate(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null")
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestCmdRun_NoneTemplate(t *testing.T) {
	RunCmdAndExpectFail(t, "openvdc", "run", "none")
}
