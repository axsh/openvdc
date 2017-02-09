// +build acceptance

package tests

import (
	"strings"
	"testing"
)

func TestOpenVDCLog(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	defer RunCmd("openvdc", "destroy", instance_id)

	RunCmdAndReportFail(t, "openvdc", "log", instance_id)
}
