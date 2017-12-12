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
      "type":"beth",
      "ipv4add": "172.16.10.10",
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
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ping -c 1 -W 3 8.8.8.8", instance_id))
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestDefaultGatewayQEMU(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_ga", tmpl)
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	// todo: find a way that does not rely on the console command for this (ssh?)
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ping -c -W 3 1 8.8.8.8", instance_id))
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
