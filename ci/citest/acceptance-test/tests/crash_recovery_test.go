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
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-stop -n "+instance_id+" -k", 10, 5)

	t.Log("Waiting for scheduler to register crash...")
	WaitConnectionStatus(t, 5*time.Minute, instance_id, "NOT_CONNECTED")

	t.Log("Re-start mesos-slave...")
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl start mesos-slave", 10, 5)
	WaitConnectionStatus(t, 5*time.Minute, instance_id, "CONNECTED")

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "stop", instance_id)
	t.Log("Waiting for instance to become STOPPED...")
	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)

	t.Log("Waiting for instance " + instance_id + " to become TERMINATED...")
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
