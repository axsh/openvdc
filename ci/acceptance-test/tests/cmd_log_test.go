// +build acceptance

package tests

import (
	"strings"
	"testing"
)

func TestOpenVDCLog(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	defer RunCmd("openvdc", "destroy", instanceID)

	RunCmdAndReportFail(t, "openvdc", "log", instanceID)
}
