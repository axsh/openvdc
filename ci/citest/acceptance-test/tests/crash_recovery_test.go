// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestCrashRecovery(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)

	//Simulate crash
	t.Log("Simulate crash by stopping mesos-slave...")
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl stop mesos-slave", 10, 5)

	//Give scheduler some time to register crash
	t.Log("Waiting for scheduler to register crash...")
	time.Sleep(2 * time.Minute)

	t.Log("Re-start mesos-slave...")
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl start mesos-slave", 10, 5)

	t.Log("Waiting for instance "+instance_id+" to recover...")
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	time.Sleep(10 * time.Second)
	t.Log("Attempting to destroy instance "+instance_id+"...")
	_, _ = RunCmdAndReportFail(t, "openvdc", "destroy", instance_id)

	t.Log("Waiting for instance "+instance_id+" to become TERMINATED...")
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
