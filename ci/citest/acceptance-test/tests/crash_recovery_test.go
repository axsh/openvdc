// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestCrashRecovery(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	stdout, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)

	//Simulate crash
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl stop mesos-slave", 10, 5)

	//Give scheduler some time to register crash
	time.Sleep(2 * time.Minute)

	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl start mesos-slave", 10, 5)
	time.Sleep(2 * time.Minute)

	stdout, _ = RunCmdAndReportFail(t, "openvdc", "stop", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	stdout, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
