// +build acceptance

package tests

import (
	"strings"
	"testing"
	"time"

	tests "github.com/axsh/openvdc/ci/citest/acceptance-test/tests"
)

var tmpl = `{
    "interfaces": [
      {
        "network_id": "Internal2",
        "type": "vmxnet3",
        "ipv4addr": "192.168.2.123"
      }
    ]
}`

var tmpl2 = `{
    "esxi_image": {
      "datastore": "datastore2",
      "name": "CentOS7_stripped"
    },
    "interfaces": [
      {
        "network_id": "Internal2",
        "type": "vmxnet3",
        "ipv4addr": "192.168.2.123"
      }
    ]
}`

func TestEsxiInstance(t *testing.T) {
	stdout, _ := tests.RunCmdAndReportFail(t, "openvdc", "run", "centos/7/esxi", tmpl)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = tests.RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	_, _ = tests.RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)

	//TODO: Use Esxi api to make sure that the vm gets properly created and destroyed.
}

func TestEsxiInstanceStripped(t *testing.T) {
	stdout, _ := tests.RunCmdAndReportFail(t, "openvdc", "run", "centos/7/esxi", tmpl2)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = tests.RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	_, _ = tests.RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	tests.WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)

}
