// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCmdReboot(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	testCmdReboot_Ubuntu14(t, instance_id)
	//testCmdReboot_Centos7(t, instance_id)

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func testCmdReboot_Ubuntu14(t *testing.T, instance_id string) {
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- sh -c 'echo \"#!/bin/sh\\ntouch /var/local/openvdc\\n\" > /etc/rc.local; chmod 755 /etc/rc.local; sync;'", instance_id), 10, 5)
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"REBOOTING"})
	WaitUntil(t, 6 * time.Minute, func() error {
		stdout, _, err := RunSsh(executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- runlevel", instance_id))
		if err != nil {
			return err
		}
		if strings.Contains(stdout.String(), "unknown") {
			time.Sleep(3 * time.Second)
			return WaitContinue
		}
		return nil
	})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- test -f /var/local/openvdc", instance_id), 10, 5)
}

func testCmdReboot_Ubuntu16(t *testing.T, instance_id string) {
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- systemctl enable rc-local.service", instance_id), 10, 5)
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- sh -c 'echo \"#!/bin/sh\\ntouch /var/local/openvdc\\n\" > /etc/rc.local; chmod 755 /etc/rc.local; sync;'", instance_id), 10, 5)
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"REBOOTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- test -f /var/local/openvdc", instance_id), 10, 5)
}

func testCmdReboot_Centos7(t *testing.T, instance_id string) {
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- chmod 755 /etc/rc.d/rc.local", instance_id), 10, 5)
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"REBOOTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- test -f /var/lock/subsys/local", instance_id), 10, 5)
}