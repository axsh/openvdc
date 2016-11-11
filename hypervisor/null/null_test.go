package null

import "testing"
import "github.com/axsh/openvdc/hypervisor"

func TestProviderRegistration(t *testing.T) {
	p, _ := hypervisor.FindProvider("null")
	if p == nil {
		t.Error("null provider is not registered.")
	}
}

func TestNullHypervisorDriver(t *testing.T) {
	n := &NullHypervisorProvider{}
	if n.Name() != "null" {
		t.Error("Invalid provider name")
	}
}
