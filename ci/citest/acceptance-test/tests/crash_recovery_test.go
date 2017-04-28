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
	_, _ = RunCmdAndReportFail(t, mesos_master_ip, "systemctl stop mesos-master")
	_, _ = RunCmdAndReportFail(t, executor_lxc_ip, "systemctl stop mesos-slave")
	_, _ = RunCmdAndReportFail(t, scheduler_ip, "systemctl stop openvdc-scheduler")

	_, _ = RunCmdAndReportFail(t, mesos_master_ip, "systemctl start mesos-master")
	_, _ = RunCmdAndReportFail(t, executor_lxc_ip, "systemctl start mesos-slave")
	_, _ = RunCmdAndReportFail(t, scheduler_ip, "systemctl start openvdc-scheduler")

	//Give mesos a moment to boot up.
	time.Sleep(10 * time.Second)

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
}
