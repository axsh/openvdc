// +build linux

package lxc

import "testing"
import "github.com/axsh/openvdc/hypervisor"

func TestProviderRegistration(t *testing.T) {
	p, _ := hypervisor.FindProvider("lxc")
	if p == nil {
		t.Error("lxc provider is not registered.")
	}
}

func TestLXCHypervisorDriver(t *testing.T) {
	p, _ := hypervisor.FindProvider("lxc")
	lxc := p.CreateDriver()
	err := lxc.CreateInstance()
	if err != nil {
		t.Error(err)
	}
	err = lxc.StartInstance()
	if err != nil {
		t.Error(err)
	}
	err = lxc.StopInstance()
	if err != nil {
		t.Error(err)
	}
	err = lxc.DestroyInstance()
	if err != nil {
		t.Error(err)
	}
}
