// +build acceptance

package tests

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"testing"
)

const zookeeper_ip = "10.0.100.10"
const mesos_master_ip = "10.0.100.11"
const scheduler_ip = "10.0.100.12"
const executor_null_ip = "10.0.100.13"
const executor_lxc_ip = "10.0.100.14"

func RunCmdThroughSsh(t *testing.T, ip string, cmd string) {
	sshConfig := &ssh.ClientConfig{
		User: "kemumaki",
		Auth: []ssh.AuthMethod{
			ssh.Password("kemumaki"),
		},
	}

	connection, err := ssh.Dial("tcp", ip+":22", sshConfig)
	if err != nil {
		t.Fatalf("SSH connection failed: %s", err.Error())
	}

	session, err := connection.NewSession()
	if err != nil {
		t.Fatalf("Unable to create a session: %s", err.Error())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		t.Logf("STDOUT:\n%s", stdout.String())
		t.Logf("STDERR:\n%s", stderr.String())

		t.Fatalf("Unable to run command: %s\n%s", cmd, err.Error())
	}

}
