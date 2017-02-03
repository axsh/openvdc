// +build acceptance

package tests

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"testing"
	"time"
)

const zookeeper_ip = "10.0.100.10"
const mesos_master_ip = "10.0.100.11"
const scheduler_ip = "10.0.100.12"
const executor_null_ip = "10.0.100.13"
const executor_lxc_ip = "10.0.100.14"

func RunSsh(ip string, cmd string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	sshConfig := &ssh.ClientConfig{
		User: "kemumaki",
		Auth: []ssh.AuthMethod{
			ssh.Password("kemumaki"),
		},
	}

	connection, err := ssh.Dial("tcp", ip+":22", sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, nil, err
	}

	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)

	return &stdout, &stderr, err
}

func RunSshAndReportFail(t *testing.T, ip string, cmd string) {
	stdout, stderr, err := RunSsh(ip, cmd)

	if err != nil {
		t.Logf("STDOUT:\n%s", stdout.String())
		t.Logf("STDERR:\n%s", stderr.String())

		t.Fatalf("Running command over ssh failed. Command: %s\n%s", cmd, err.Error())
	}
}

func RunSshWithTimeoutAndReportFail(t *testing.T, ip string, cmd string, tries int, sleeptime time.Duration) {
	tried := 1
	stdout, stderr, err := RunSsh(ip, cmd)

	for err != nil {
		if tried >= tries {
			t.Logf("STDOUT:\n%s", stdout.String())
			t.Logf("STDERR:\n%s", stderr.String())

			t.Fatalf("Running command over ssh failed. Tried %s times.\nCommand: %s\n%s", tries, cmd, err.Error())
		}

		time.Sleep(sleeptime * time.Second)

		tried += 1
		stdout, stderr, err = RunSsh(ip, cmd)
	}
}
