package local

import (
	"testing"
)

func TestNewLocalResourceCollector(t *testing.T) {
	c, err := NewLocalResourceCollector()
	if c == nil && err != nil {
		t.Errorf("Unable to create local resource collector.")
	}
}
