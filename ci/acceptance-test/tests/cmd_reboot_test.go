// +build acceptance

package tests

import (
	"strings"
	"testing"
	"os/exec"
	"time"

	"github.com/tidwall/gjson"
)

func TestCmdReboot(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	waitUntil(t, 5 * time.Minute, func()error {
		cmd := exec.Command("openvdc", "show", instance_id)
		buf, err := cmd.CombinedOutput()
		if err !=nil {
			return err
		}
		result := gjson.GetBytes(buf, "instance.lastState.state")
		if result.String() == "RUNNING" {
			return nil
		}
		time.Sleep(5 * time.Second)
		return WaitContinue
	})

  // This test scenario is only valid for CentOS 7 /etc/rc.d/rc.local.
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-execute -n %s -- chmod 755 /etc/rc.d/rc.local", instance_id), 10, 5)
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, fmt.Sprintf("sudo lxc-execute -n %s -- test -f /var/lock/subsys/local", instance_id), 10, 5)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
}
