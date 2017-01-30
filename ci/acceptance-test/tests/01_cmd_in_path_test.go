// +build acceptance

package tests

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestOpenVDCCmdInPath(t *testing.T) {
	var out bytes.Buffer

	cmd := exec.Command("openvdc")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		t.Fatal("Unable to run openvdc command: " + err.Error())
	}

	if !strings.HasPrefix(out.String(), "Usage:") {
		t.Fatal("Running openvdc without arguments didn't print usage. Instead got: " + out.String())
	}
}
