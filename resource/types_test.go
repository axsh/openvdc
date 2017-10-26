package resource

import (
	"testing"

	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
)

type testResourceCollector struct {}
func (c *testResourceCollector) GetCpu() (*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetMem() (*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetDisk() ([]*model.Resource, error) { return nil, nil }
func (c *testResourceCollector) GetLoadAvg() (*model.LoadAvg, error) { return nil, nil }

func NewTestCollector(conf *viper.Viper) (ResourceCollector, error) {
	return &testResourceCollector{}, nil
}

func TestRegisterCollector (t *testing.T)  {
	c := RegisterCollector("test", NewTestCollector)
	if c != nil {
		t.Errorf("Unable to register resource collector.")
	}
}

func TestNewCollector (t *testing.T) {
	viper.SetDefault("resource-collector.mode", "test")
	rc, err := NewCollector(viper.GetViper())

	if rc == nil || err != nil {
		t.Errorf("Unable to create new resource collector.")
	}
}
