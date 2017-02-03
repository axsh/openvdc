// +build acceptance

package tests

import (
	"strings"
	"testing"
)

func TestOpenVDCCmdInPath(t *testing.T) {
	stdout, _ := RunCmd(t, "openvdc")

	if !strings.HasPrefix(stdout.String(), "Usage:") {
		t.Fatal("Running openvdc without arguments didn't print usage. Instead got: " + stdout.String())
	}
}
