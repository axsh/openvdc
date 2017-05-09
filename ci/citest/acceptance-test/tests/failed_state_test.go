// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestFailedState_StartInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "start"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"QUEUED", "STARTING"})
}

func TestFailedState_CreateInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "create"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"QUEUED", "STARTING"})
}

func TestFailedState_StopInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "stop"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "stop", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"STOPPING"})
}

func TestFailedState_RebootInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "reboot"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "reboot", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"REBOOTING"})
}

func TestFailedState_DestroyInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/null", `{"crash_stage": "destroy"}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunCmdAndReportFail(t, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "FAILED", []string{"SHUTTINGDOWN"})
}
