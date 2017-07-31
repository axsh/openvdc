// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"

	tests "github.com/axsh/openvdc/ci/citest/acceptance-test/tests"
)

func TestEsxiInstance(t *testing.T) {
	stdout, _ := tests.RunCmdAndReportFail(t, "openvdc", "run", "centos/7/esxi", `{"interfaces":[{"Ipv4Addr":"192.168.2.123"}]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = tests.RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	_, _ = tests.RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)

	//TODO: Use Esxi api to make sure that the vm gets properly created and destroyed.
}
