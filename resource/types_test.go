package resource

import (
	"testing"

	"github.com/axsh/openvdc/model"
)

type testResourceCollector struct {}
func (c *testResourceCollector) GetCpu() (*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetMem() (*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetDisk() ([]*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetLoadAvg() (*model.LoadAvg, error) { return nil, nil }

func NewTestCollector() (ResourceCollector, error) {
	return &testResourceCollector{
	}, nil
}

func TestRegisterCollector (t *testing.T)  {
	c := RegisterCollector("test", NewTestCollector)
	if c != nil {
		t.Errorf("Unable to register resource collector.")
	}
}

func TestNewCollector (t *testing.T) {
	rc, err := NewCollector("test")

	if rc == nil || err != nil {
		t.Errorf("Unable to create new resource collector.")
	}
}
