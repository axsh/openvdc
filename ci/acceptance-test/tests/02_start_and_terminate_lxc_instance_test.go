// +build acceptance

package tests

import (
	"fmt"
	"strings"
	"testing"
	"os/exec"
	"time"

	"github.com/tidwall/gjson"
)

var WaitContinue = fmt.Errorf("continue")

func waitUntil(t *testing.T, d time.Duration, cb func()error) {
	startAt := time.Now()
	var err error
	for {
		err = cb()
		if err == WaitContinue {
			if time.Now().Sub(startAt) > d {
				err = fmt.Errorf("Timed out for %d sec", d / time.Second)
				break
			}
			continue
		}
		break
	}

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestLXCInstance(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/lxc")
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)

	cmd := exec.Command("openvdc", "show", instance_id)
	waitUntil(t, 5 * time.Minute, func()error {
		buf, err := cmd.CombinedOutput()
		if err !=nil {
			return err
		}
		result := gjson.GetBytes(buf, "instance.lastState.state")
		if result.String() == "RUNNING" {
			return nil
		}
		time.Sleep(5 * time.Second)
		return WaitContinue
	})

	//TODO: Wait for instance state RUNNING in OpenVDC before we do this
	//This will require us to unmarshall json from the 'openvdc show' command
	RunSshWithTimeoutAndReportFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
	//TODO: Run only once after we've waited for RUNNING
	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)

	//TODO: Don't rely on the standard 'command failed' error.
	//It's more clear to say "container didn't get destroyed after calling openvdc destroy"
	RunSshWithTimeoutAndExpectFail(t, executor_lxc_ip, "sudo lxc-info -n "+instance_id, 10, 5)
}
