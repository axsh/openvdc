// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func runConsoleCmd(instance_id string, t *testing.T) {
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("echo 'ls' | openvdc console %s", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("echo 'false' | openvdc console %s", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ls", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- ls", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- false", instance_id))
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestLXCCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	runConsoleCmd(instance_id, t)
}

func TestKVMCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/kvm")
	instance_id := strings.TrimSpace(stdout.String())
	runConsoleCmd(instance_id, t)
}
