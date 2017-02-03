// +build acceptance

package tests

import (
	"testing"
)

func TestLXCInstance(t *testing.T) {
	instance_id, _ := RunCmd(t, "openvdc", "run", "centos/7/lxc")

	RunCmdThroughSsh(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id.String())
}
