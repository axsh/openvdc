// +build acceptance

package tests

import (
	"testing"
)

func TestLXCInstance(t *testing.T) {
	_, _ = RunCmd(t, "openvdc", "run", "centos/7/lxc")
}
