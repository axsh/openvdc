// +build acceptance

package tests

import (
	"strings"
	"fmt"
	"testing"
	"time"
)

func TestQEMUInstance_KVMDisabled(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
<<<<<<< HEAD
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_qemu_ip, "echo info name | sudo ncat -U /var/openvdc/qemu-instances/"+instance_id+"/monitor.socket", 10, 5)
	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestQEMUInstance_KVMEnabled(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_kvm", `{"interfaces":[{"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	_, _ = RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_qemu_ip, "echo info name | sudo ncat -U /var/openvdc/qemu-instances/"+instance_id+"/monitor.socket", 10, 5)
	_, _ = RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestQEMUInstance_LinuxBrNICx2(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_kvm",
		`{"interfaces":[{"type":"veth"}, {"type":"veth"}], "node_groups":["linuxbr"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_qemu_ip, "echo info name | sudo ncat -U /var/openvdc/qemu-instances/"+instance_id+"/monitor.socket", 10, 5)
	stdout, _, err := RunSsh(executor_qemu_ip, fmt.Sprintf("/usr/sbin/bridge link show dev %s", instance_id+"_00"))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Interface %s is not attached", instance_id+"_00")
	} else {
		if testing.Verbose() {
			t.Log("bridge link show dev "+instance_id+"_00: ", stdout.String())
		}
	}
	stdout, _, err = RunSsh(executor_qemu_ip, fmt.Sprintf("/usr/sbin/bridge link show dev %s", instance_id+"_01"))
	if err != nil {
		t.Error(err)
	}
	if stdout.Len() == 0 {
		t.Errorf("Interface %s is not attached", instance_id+"_01")
	} else {
		if testing.Verbose() {
			t.Log("bridge link show dev "+instance_id+"_01: ", stdout.String())
		}
	}
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
}

func TestQEMUInstance_OVSBrNICx2(t *testing.T) {
	stdout, _ := RunCmdAndReportFail(t, "openvdc", "run", "centos/7/qemu_kvm",
		`{"interfaces":[{"type":"vif"}, {"type":"vif"}], "node_groups":["ovs"]}`)
	instance_id := strings.TrimSpace(stdout.String())

	RunCmdAndReportFail(t, "openvdc", "show", instance_id)
	WaitInstance(t, 10*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})
	RunSshWithTimeoutAndReportFail(t, executor_qemu_ovs_ip, "echo info name | sudo ncat -U /var/openvdc/qemu-instances/"+instance_id+"/monitor.socket", 10, 5)
	stdout, _, err := RunSsh(executor_qemu_ovs_ip, fmt.Sprintf("sudo /usr/bin/ovs-vsctl port-to-br %s", instance_id+"_00"))

	if err != nil {
		t.Error(err)
	} else {
		if testing.Verbose() {
			t.Log("ovs-vsctl port-to-br "+instance_id+"_00", stdout.String())
		}
	}
	stdout, _, err = RunSsh(executor_qemu_ovs_ip, fmt.Sprintf("sudo /usr/bin/ovs-vsctl port-to-br %s", instance_id+"_01"))
	if err != nil {
		t.Error(err)
	} else {
		if testing.Verbose() {
			t.Log("ovs-vsctl port-to-br "+instance_id+"_01", stdout.String())
		}
	}
	RunCmdWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
=======
	WaitInstance(t, 5*time.Minute, instance_id, "RUNNING", []string{"QUEUED", "STARTING"})

	// maybe we should open the server on a port rather than as file?
	// 	stdout, err := RunSshWithTimeoutAndReportFail(t, executor_qemu_ip, "echo info name | nc localhost /var/lib/openvdc/instances/"+instance_id+".monitor", 10, 5)
	// 	_, _ = RunSshWithTimeoutAndReportFail(t, 10, 5, "openvdc", "destroy", instance_id)
	// 	WaitInstance(t, 5*time.Minute, instance_id, "TERMINATED", nil)
>>>>>>> 25c0065cfee5d5ebbd073858ac5c24ca4f08e65d
}
