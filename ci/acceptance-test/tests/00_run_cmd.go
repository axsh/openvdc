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
		t.Log("STDOUT:\n" + stdout.String())
		t.Log("STDERR:\n" + stderr.String())

		t.Fatal("Unable to run " + name + " command: " + err.Error())
	}

	t.Logf("%+v\n", stdout)

	return &stdout, &stderr
}
