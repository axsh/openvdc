package esxi

import (
	"testing"
	"github.com/spf13/viper"
)

func TestNewEsxiResourceCollector(t *testing.T) {
	viper.SetDefault("resource-collector.mode", "esxi")
	viper.SetDefault("hypervisor.esxi-user", "test")
	viper.SetDefault("hypervisor.esxi-pass", "test")
	viper.SetDefault("hypervisor.esxi-ip", "0.0.0.0")
	viper.SetDefault("hypervisor.esxi-datacenter", "test")
	viper.SetDefault("hypervisor.esxi-datastore", "test")

	c, err := NewEsxiResourceCollector(viper.GetViper()) 
	if c == nil && err != nil {
		t.Errorf("Unable to create esxi resource collector.")
	}
}
