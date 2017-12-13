// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var tmpl = `{
  "interfaces": [
    {
      "type":"veth",
      "ipv4addr": "172.16.10.100",
      "gateway": "172.16.10.1"
    }
  ],
  "node_groups":["linuxbr"]
}`

func TestDefaultGatewayLXC(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", tmpl)
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	// todo: find a way that does not rely on the console command for this (ssh?)
	stdout, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "sh", "-c",
		fmt.Sprintf("openvdc console %s '%s'", instance_id, "ping -c 1 -W 3 8.8.8.8"))
	t.Log("\n", stdout.String())

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestDefaultGatewayQEMU(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_ga", tmpl)
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	// todo: find a way that does not rely on the console command for this (ssh?)
	stdout, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "sh", "-c",
		fmt.Sprintf("openvdc console %s '%s'", instance_id, "ping -c 1 -W 3 8.8.8.8"))
	t.Log("\n", stdout.String())

	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
