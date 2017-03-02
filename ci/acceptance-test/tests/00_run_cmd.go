// +build acceptance

package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/tidwall/gjson"
)

func RunCmd(name string, arg ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return &stdout, &stderr, err
}

func RunCmdAndReportFail(t *testing.T, name string, arg ...string) (*bytes.Buffer, *bytes.Buffer) {
	stdout, stderr, err := RunCmd(name, arg...)

	if err != nil {
		t.Logf("STDOUT:\n%s", stdout.String())
		t.Logf("STDERR:\n%s", stderr.String())

		t.Fatalf("Unable to run command: '%s %v'\n%s", name, arg, err.Error())
	}

	return stdout, stderr
}

func RunCmdAndExpectFail(t *testing.T, name string, arg ...string) (*bytes.Buffer, *bytes.Buffer) {
	stdout, stderr, err := RunCmd(name, arg...)

	if err == nil {
		t.Logf("STDOUT:\n%s", stdout.String())
		t.Logf("STDERR:\n%s", stderr.String())

		t.Fatalf("Expected Command to fail: '%s %v'\n%s", name, arg, err.Error())
	}

	return stdout, stderr
}

func RunCmdWithTimeoutAndReportFail(t *testing.T, tries int, sleeptime time.Duration, name string, arg ...string) (*bytes.Buffer, *bytes.Buffer) {
	tried := 1
	stdout, stderr, err := RunCmd(name, arg...)

	for err != nil {
		if tried >= tries {
			t.Logf("STDOUT:\n%s", stdout.String())
			t.Logf("STDERR:\n%s", stderr.String())

			//TODO: Improve this log by including arg
			t.Fatalf("Running '%s' failed. Tried %d times.", name, tries)
		}

		time.Sleep(sleeptime * time.Second)

		tried += 1
		stdout, stderr, err = RunCmd(name, arg...)
	}

	return stdout, stderr
}

func RunCmdWithTimeoutAndExpectFail(t *testing.T, tries int, sleeptime time.Duration, name string, arg ...string) (*bytes.Buffer, *bytes.Buffer) {
	tried := 1
	stdout, stderr, err := RunCmd(name, arg...)

	for err == nil {
		if tried >= tries {
			t.Logf("STDOUT:\n%s", stdout.String())
			t.Logf("STDERR:\n%s", stderr.String())

			t.Fatalf("Expected '%s' to fail. Tried %d times.", name, tries)
		}

		time.Sleep(sleeptime * time.Second)

		tried += 1
		stdout, stderr, err = RunCmd(name, arg...)
	}

	return stdout, stderr
}

var WaitContinue = fmt.Errorf("continue")

func WaitUntil(t *testing.T, d time.Duration, cb func() error) {
	startAt := time.Now()
	var err error
	for {
		err = cb()
		if err == WaitContinue {
			if time.Now().Sub(startAt) > d {
				err = fmt.Errorf("Timed out for %d sec", d/time.Second)
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

func WaitInstance(t *testing.T, d time.Duration, instanceID string, goalState string, interimStates []string) {
	WaitUntil(t, d, func() error {
		cmd := exec.Command("openvdc", "show", instanceID)
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		result := gjson.GetBytes(buf, "instance.lastState.state")
		if result.String() == "RUNNING" {
			return nil
		} else if interimStates != nil {
			for _, state := range interimStates {
				if result.String() != state {
					return fmt.Errorf("Unexpected Instance State: %s goal=%s found=%s", instanceID, goalState, result.String())
				}
			}
		}
		time.Sleep(5 * time.Second)
		return WaitContinue
	})
}
