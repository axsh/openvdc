// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"
)

func TestKVMInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/kvm", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	// maybe we should open the server on a port rather than as file?
	stdout, err := RunSshWithTimeoutAndReportFail(t, executor_kvm_ip, "echo info name | nc localhost /var/lib/openvdc/instances/"+instance_id+".monitor", 10, 5)
	_, _ = RunSshWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
