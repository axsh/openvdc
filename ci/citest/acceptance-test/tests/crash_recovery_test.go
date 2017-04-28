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

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)

	//Simulate crash
	RunSshWithTimeoutAndReportFail(t, mesos_master_ip, "sudo systemctl stop mesos-master", 10, 5)
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl stop mesos-slave", 10, 5)
	RunSshWithTimeoutAndReportFail(t, scheduler_ip, "sudo systemctl stop openvdc-scheduler", 10, 5)

	RunSshWithTimeoutAndReportFail(t, mesos_master_ip, "sudo systemctl start mesos-master", 10, 5)
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo systemctl start mesos-slave", 10, 5)
	RunSshWithTimeoutAndReportFail(t, scheduler_ip, "sudo systemctl start openvdc-scheduler", 10, 5)

	//Give mesos a moment to boot up.
	time.Sleep(10 * time.Second)

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "stop", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
