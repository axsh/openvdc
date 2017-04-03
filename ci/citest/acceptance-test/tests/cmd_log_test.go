// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestOpenVDCLog(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	defer RunCmd("openvdc", "destroy", instance_id)

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "log", instance_id)
}
