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
	// TODO: For CentOS7, just needs to set 755 to the file. Then you'll see /var/lock/subsys/local with recent timestamp.
	// RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- chmod 755 /etc/rc.d/rc.local", instance_id), 10, 5)
	// TODO: /etc/rc.local location is for ubuntu. Needs rewrite for CentOS7
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- sh -c 'echo \"#!/bin/sh\ntouch /var/tmp/local\" >> /etc/rc.local'; sudo chmod 755 /etc/rc.local;", instance_id), 10, 5)
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"REBOOTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- test -f /var/tmp/local", instance_id), 10, 5)
	//RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-attach -n %s -- test -f /var/lock/subsys/local", instance_id), 10, 5)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
