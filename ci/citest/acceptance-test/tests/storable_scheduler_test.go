// +build acceptance

package tests

import (
	"testing"
)

func TestStorable_scheduler_LxcHugeTemplate(t *testing.T) {
	_, stderr := RunCmdAndExpectFail(t, "openvdc", "run", "centos/7/lxc_huge")
	if stderr.String() != "There is no machine can satisfy resource requirement" {
		t.Error("There is no machine can satisfy resource requirement but work")
	}
}
