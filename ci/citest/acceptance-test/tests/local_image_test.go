// +build acceptance

package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalImage(t *testing.T) {

	// Use custom lxc-template.
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "https://raw.githubusercontent.com/axsh/openvdc/template-misc/templates/centos/7/lxc2.json", `{"node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)

	configFile := filepath.Join("/var/lib/lxc/", instance_id, "config")

	stdout, _, err := RunSsh(executor_lxc_ip, fmt.Sprintf("echo $(head -n 1 %s)", configFile))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Couldn't read %s", configFile)
	}

	s := strings.Split(stdout.String(), "/")
	templateUsed := s[len(s)-1]
	if templateUsed != "" || templateUsed != "lxc-openvdc" {
		t.Errorf("Expected templateUsed to be 'lxc-openvdc', got:  %s", templateUsed)
	}

	_, _ = RunCmdAndReportFail(t, "openvdc", "stop", instance_id)

	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
