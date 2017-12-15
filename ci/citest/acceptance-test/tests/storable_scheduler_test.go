// +build acceptance

package tests

import (
	"testing"
)

func TestStorable_scheduler_NullTemplate(t *testing.T) {
	_, stderr := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc_huge")
	if stderr.String() != "There is no machine can satisfy resource requirement" {
		t.Error("There is no machine can satisfy resource requirement but work")
	}

	// instance_id := strings.TrimSpace(stdout.String())

	// WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	// RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	// WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
