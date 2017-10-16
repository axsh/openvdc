package resources

import (
	"fmt"
	"strings"

	"github.com/axsh/openvdc/model"
)

type ResourceCollector interface {
	GetCpu() (*model.Resource, error)
	GetMem() (*model.Resource, error)
	GetDisk() ([]*model.Resource, error)
	GetLoadAvg() (*model.LoadAvg, error)
}

type collectorType func() (ResourceCollector, error)

var (
	collectors = make(map[string]collectorType)
)

func NewCollector(name string) (ResourceCollector, error) {
	collector, exists := collectors[name]
	if !exists {
		knownCollecotrs := make([]string, len(collectors))
		for c, _ := range collectors {
			knownCollecotrs = append(knownCollecotrs, c)
		}
		return nil, fmt.Errorf("Failed getCollector() Must be one of: %s",
			strings.Join(knownCollecotrs, ", "))
	}

	return collector()
}

func RegisterCollector(name string, collectorType collectorType) error {
	if collectorType == nil {
		return fmt.Errorf("Unknown resource collector: %s", name)
	}
	if _, exists := collectors[name]; exists {
		return fmt.Errorf("Duplicate resource collector registration: %s", name)
	}
	collectors[name] = collectorType
	return nil
}
