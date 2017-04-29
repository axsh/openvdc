// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestFailedState_StartInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "start"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"QUEUED", "STARTING"})
}
