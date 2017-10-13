// +build acceptance

package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func runConsoleCmdPiped(instance_id string, t *testing.T) {
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("echo 'ls' | openvdc console %s", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("echo 'false' | openvdc console %s", instance_id))
}

func runConsoleCmd(instance_id string, t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "console", instance_id, "--show")
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s ls", instance_id))
	RunCmdAndReportFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- ls", instance_id))
	RunCmdAndExpectFail(t, "sh", "-c", fmt.Sprintf("openvdc console %s -- false", instance_id))
}

func runConsoleCmdWithPrivatekey(instance_id string, private_key_path string, t *testing.T, expect_fail bool) {
	if expect_fail {
		RunCmdAndExpectFail(t, "openvdc", "console", instance_id, "-i", private_key_path)
	} else {
		RunCmdAndReportFail(t, "openvdc", "console", instance_id, "-i", private_key_path)
	}
}

func TestCmdConsole_ShowOptionAuthenticationNone(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"authentication_type":"none"}`)
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmd(instance_id, t)
	runConsoleCmdPiped(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestLXCCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmd(instance_id, t)
	runConsoleCmdPiped(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestLXCCmdConsole_AuthenticationPubkey(t *testing.T) {
	// Make key pair by ssh-keygen
	private_key_path := "./testRsa"
	private_key_path2 := "./testRsa2"
	_, _, err := RunCmd("ssh-keygen", "-t", "rsa", "-f", private_key_path, "-C", "", "-N", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	_, _, err = RunCmd("ssh-keygen", "-t", "rsa", "-f", private_key_path2, "-C", "", "-N", "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Read public key
	data, err := ioutil.ReadFile(private_key_path + ".pub")
	if err != nil {
		t.Fatalf("Can not read public key: %s\n", err.Error())
	}
	public_key := strings.Replace(string(data), "\n", "", -1)
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc", `{"authentication_type":"pub_key","ssh_public_key":"`+public_key+`"}`)

	// runConsole()
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmdWithPrivatekey(instance_id, private_key_path, t, false)
	runConsoleCmdWithPrivatekey(instance_id, private_key_path2, t, true) // This can not be authenticated.
	//vrunConsoleCmdPiped(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestQEMUCmdConsole_ShowOption(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_ga")
	instance_id := strings.TrimSpace(stdout.String())
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	runConsoleCmd(instance_id, t)
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "TERMINATED", nil)
}
