package esxi

import (
	"testing"
)

func TestNewEsxiResourceCollector(t *testing.T) {
	c, err := NewEsxiResourceCollector() 
	if c == nil && err != nil {
		t.Errorf("Unable to create esxi resource collector.")
	}
}
