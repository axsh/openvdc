// +build acceptance

package tests

import (
	"fmt"
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

func TestLXCInstance_NICx2(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc",
		`{"interfaces":[{"type":"veth", "bridge":"linux"}, {"type":"veth", "bridge":"linux"}]}`)
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
	stdout, _, err := RunSsh(executor_lxc_ip, fmt.Sprintf("bridge link show dev %s", instance_id+"_00"))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Interface %s is not attached", instance_id+"_00")
	} else {
		if testing.Verbose() {
			t.Log("bridge link show dev "+instance_id+"_00: ", stdout.String())
		}
	}
	stdout, _, err = RunSsh(executor_lxc_ip, fmt.Sprintf("bridge link show dev %s", instance_id+"_01"))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Interface %s is not attached", instance_id+"_01")
	} else {
		if testing.Verbose() {
			t.Log("bridge link show dev "+instance_id+"_01: ", stdout.String())
		}
	}

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
