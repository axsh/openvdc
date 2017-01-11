// +build acceptance

package tests

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"strings"
	"testing"
)

func TestOpenVDCCmdInPath(t *testing.T) {
	sshConfig := &ssh.ClientConfig{
		User: "kemumaki",
		Auth: []ssh.AuthMethod{
			ssh.Password("kemumaki"),
		},
	}

	connection, err := ssh.Dial("tcp", "10.0.100.13:22", sshConfig)
	if err != nil {
		t.Fatal("SSH connection failed: " + err.Error())
	}

	session, err := connection.NewSession()
	if err != nil {
		t.Fatal("Unable to create a session: " + err.Error())
	}

	var out bytes.Buffer
	session.Stdout = &out

	err = session.Run("openvdc")
	if err != nil {
		t.Fatal("Unable to run openvdc command: " + err.Error())
	}

	if !strings.HasPrefix(out.String(), "Usage:") {
		t.Fatal("Running openvdc without arguments didn't print usage. Instead got: " + out.String())
	}
}
