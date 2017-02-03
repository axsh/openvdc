// +build acceptance

package tests

import (
	"bytes"
	"os/exec"
	"testing"
	"time"
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
