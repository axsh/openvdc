// +build acceptance

package tests

import (
	"bytes"
	"os/exec"
	"testing"
)

func RunCmd(t *testing.T, name string, arg ...string) (*bytes.Buffer, *bytes.Buffer) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("STDOUT:\n%s", stdout.String())
		t.Logf("STDERR:\n%s", stderr.String())

		t.Fatalf("Unable to run command: '%s %v'\n%s", name, arg, err.Error())
	}

	t.Logf("%+v\n", stdout)

	return &stdout, &stderr
}
