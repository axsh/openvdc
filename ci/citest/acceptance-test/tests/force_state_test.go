// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestForceState(t *testing.T) {
        stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
        instance_id := strings.TrimSpace(stdout.String())

        WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	RunCmdAndReportFail(t, "openvdc", "forcestate", instance_id, "stopped")
	WaitInstance(t, 1*time.Minute, instance_id, "STOPPED", nil)

	RunCmdAndReportFail(t, "openvdc", "forcestate", instance_id, "terminated")
        WaitInstance(t, 1*time.Minute, instance_id, "TERMINATED", nil)

	RunCmdAndReportFail(t, "openvdc", "forcestate", instance_id, "running")
        WaitInstance(t, 1*time.Minute, instance_id, "RUNNING", nil)

        _, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
        t.Log("Waiting for instance " + instance_id + " to become TERMINATED...")
        WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
