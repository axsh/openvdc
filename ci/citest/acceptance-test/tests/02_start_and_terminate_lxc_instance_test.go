// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestLXCInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"interfaces":[{"type":"veth", "bridge":"linux"}]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
	//TODO: Run only once after we've waited for RUNNING
	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)

	//TODO: Don't rely on the standard 'command failed' error.
	//It's more clear to say "container didn't get destroyed after calling openvdc destroy"
	RunSshWithTimeoutAndExpectFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
}
