// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"

	tests "github.com/axsh/openvdc/ci/citest/acceptance-test/tests"
)

func TestEsxiCmdConsole_ShowOption(t *testing.T) {
    stdout, _ := tests.RunCmdAndReportFail(t, "openvdc", "run", "centos/7/esxi")
    instance_id := strings.TrimSpace(stdout.String())
    tests.WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

    time.Sleep(time.Second*3)
    tests.RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
    tests.RunCmdAndReportFail(t, "sh", "-c", strings.Join([]string{"openvdc", "console", instance_id, "ls"}, " "))
    tests.RunCmdAndReportFail(t, "sh", "-c", strings.Join([]string{"openvdc", "console", instance_id, "--", "ls"}, " "))
    tests.RunCmdAndExpectFail(t, "sh", "-c", strings.Join([]string{"openvdc", "console", instance_id, "--", "false"}, " "))

    tests.RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
    tests.WaitInstance(t, 10*time.Minute, instance_id, "TERMINATED", nil)
}
