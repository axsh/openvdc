package main

import (
	"fmt"
	"os"
	"time"
	"strings"

	"github.com/axsh/openvdc/model"
)

type ResourceCollector interface {
	GetCpu()     (*model.Resource, error)
	GetMem()     (*model.Resource, error)
	GetDisk()    ([]*model.Resource, error)
	GetLoadAvg() (*model.LoadAvg, error)
}

type VDCAgent struct {
	collector ResourceCollector
	resources *model.ComputingResources
	AgentId   string
}

var (
	updateInteval time.Duration = 5
	collectors = make(map[string]collectorType)
	// to be received from configuration file or as flag
	tmpCollectorType = "esxi"
)

type collectorType func () (ResourceCollector, error)

func newCollector(name string) (ResourceCollector, error) {
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

func registerCollector(name string, collectorType collectorType) error {
	if collectorType == nil {
		return fmt.Errorf("Unknown resource collector: %s", name)
	}
	if _, exists := collectors[name]; exists {
		return fmt.Errorf("Duplicate resource collector registration: %s", name)
	}
	collectors[name] = collectorType
	return nil
}

func init() {
	registerCollector("esxi", NewEsxiResourceCollector)
	registerCollector("local", NewLocalResourceCollector)
}

func main() {
	c, err := newCollector(tmpCollectorType);
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	agent := VDCAgent{
		collector: c,
		resources: &model.ComputingResources{},
	}
	agent.GetResources()

	if err := agent.Run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func (a *VDCAgent) Run() error {
	for {
		if err := a.GetResources(); err != nil {
			return err
		}
		fmt.Println(a.resources)
		time.Sleep(time.Second * updateInteval)
	}
}

func (a *VDCAgent) GetResources() error {
	var err error

	if a.resources.Cpu, err = a.collector.GetCpu(); err != nil {
		return err
	}
	if a.resources.Memory, err = a.collector.GetMem(); err != nil {
		return err
	}
	if a.resources.Storage, err = a.collector.GetDisk(); err != nil {
		return err
	}
	if a.resources.Load, err = a.collector.GetLoadAvg(); err != nil {
		return err
	}
	return nil
}
