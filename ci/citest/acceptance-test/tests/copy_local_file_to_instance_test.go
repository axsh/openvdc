// +build acceptance

package tests

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCopyLocalFileToInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)

	testFileContent := "Hi, this is a test."
	testFileFolder := "/tmp/"
	testFileName := "testfile"
	testFilePath := filepath.Join(testFileFolder, testFileName)

	err := ioutil.WriteFile(testFilePath, []byte(testFileContent), 0644)
	if err != nil {
		t.Error(err)
	}
	RunCmdAndReportFail(t, "openvdc", "copy", testFilePath, instance_id+":"+testFileFolder)
	copiedFile := filepath.Join("/var/lib/lxc/", instance_id, "/rootfs/", testFilePath)
	stdout, _, err = RunSsh(executor_lxc_ip, fmt.Sprintf("sudo cat %s", copiedFile))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Couldn't read %s", copiedFile)
	}
	if stdout.String() != testFileContent {
		t.Errorf("Content mismatch. Got: %s, expected: %s", stdout.String(), testFileContent)
	}

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "stop", instance_id)
	t.Log("Waiting for instance to become STOPPED...")
	WaitInstance(t, 5*time.Minute, instance_id, "STOPPED", []string{"RUNNING", "STOPPING"})

	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)

	t.Log("Waiting for instance " + instance_id + " to become TERMINATED...")
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}
