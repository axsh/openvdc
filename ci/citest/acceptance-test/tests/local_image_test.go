// +build acceptance

package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	if err := RestoreAssets("/var/tmp", "fixtures"); err != nil {
		panic(err)
	}
}

func TestLocalImage(t *testing.T) {

	stdout, _, err := RunSsh(scheduler_ip, fmt.Sprintf("[ -f /var/www/html/images/centos/7/amd64/meta.tar.xz ] && echo meta.tar.xz found || echo meta.tar.xz not found"))

	if err != nil {
		t.Error(err)
	}

	t.Log(stdout.String())

	// Use custom lxc-template.
	stdout, _ = RunCmdAndReportFail(t, "openvdc", "run", "/var/tmp/fixtures/lxc2.json", `{"node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	configFile := filepath.Join("/var/lib/lxc/", instance_id, "config")
	stdout, _, err = RunSsh(executor_lxc_ip, fmt.Sprintf("echo | sudo head -n 1 %s", configFile))

	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Couldn't read %s", configFile)
	}

	if testing.Verbose() {
		t.Log(stdout.String())
	}

	s := strings.Split(strings.TrimSpace(stdout.String()), "/")
	templateUsed := s[len(s)-1]
	if templateUsed != "lxc-openvdc" {
		t.Errorf("Expected templateUsed to be 'lxc-openvdc', got:  %s", templateUsed)
	}

	_, _ = RunCmdAndReportFail(t, "openvdc", "stop", instance_id)

	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
