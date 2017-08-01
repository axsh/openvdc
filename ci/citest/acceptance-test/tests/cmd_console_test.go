// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func runConsoleCmdPiped(instance_id string, t *testing.T) {
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("echo 'ls' | openvdc console %s", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("echo 'false' | openvdc console %s", instance_id))
}

func runConsoleCmd(instance_id string, t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ls", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- ls", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- false", instance_id))
}

func TestCmdConsole_Authentication(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"authentication_type":"none"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestLXCCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmd(instance_id, t)
	runConsoleCmdPiped(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestQEMUCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_ga")
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmd(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "TERMINATED", nil)
}
