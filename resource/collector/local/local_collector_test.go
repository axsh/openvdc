package local

import (
	"testing"
	"github.com/spf13/viper"
)

func TestNewLocalResourceCollector(t *testing.T) {
	viper.SetDefault("resource-collector.mode", "local")
	c, err := NewLocalResourceCollector(viper.GetViper())
	if c == nil && err != nil {
		t.Errorf("Unable to create local resource collector.")
	}
}
