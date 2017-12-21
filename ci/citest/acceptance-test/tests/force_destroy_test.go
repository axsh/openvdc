// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestForceDestroy(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	//Simulate stuck state
	RunCmdAndReportFail(t, "openvdc", "force-state", instance_id, "starting")
	WaitInstance(t, 1*time.Minute, instance_id, "STARTING", nil)

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", "--force", instance_id)
	t.Log("Waiting for instance " + instance_id + " to become TERMINATED...")
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
