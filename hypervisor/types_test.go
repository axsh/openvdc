package hypervisor

import (
	"testing"

	"github.com/axsh/openvdc/model"
)

type testProvider struct{}

func (p *testProvider) Name() string {
	return "test"
}

func (p *testProvider) CreateDriver(string) (HypervisorDriver, error) {
	return &testDriver{}, nil
}

type testDriver struct{}

func (d *testDriver) StartInstance() error {
	return nil
}

func (d *testDriver) StopInstance() error {
	return nil
}

func (d *testDriver) RebootInstance() error {
	return nil
}

func (d *testDriver) InstanceConsole() error {
	return nil
}

func (d *testDriver) CreateInstance(*model.Instance, model.ResourceTemplate) error {
	return nil
}

func (d *testDriver) DestroyInstance() error {
	return nil
}

func TestProviderRegistry(t *testing.T) {
	{
		RegisterProvider("test", &testProvider{})
		p, _ := FindProvider("test")
		if p == nil {
			t.Errorf("Could not find test provider.")
		}
	}

	{
		p, _ := FindProvider("unknown")
		if p != nil {
			t.Error("Fails to detect the provider does not exist.")
		}
	}
}
