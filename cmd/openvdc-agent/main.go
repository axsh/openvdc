package main

import (
	"fmt"
	"os"
	"time"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
)

type VDCAgent struct {
	collector resources.ResourceCollector
	resources *model.ComputingResources
	AgentId   string
}

var (
	updateInteval time.Duration = 5
	// TODO: receive from configuration file or as flag
	tmpCollectorType = "local"
)

func main() {
	var agent *VDCAgent

	if c, err := resources.NewCollector(tmpCollectorType);err != nil {
		fmt.Println(err)
		os.Exit(-1)
	} else {
		agent = newVDCAgent(c)
	}
	agent.GetResources()

	if err := agent.Run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func newVDCAgent(c resources.ResourceCollector) *VDCAgent {
	return &VDCAgent{
		collector: c,
		resources:  &model.ComputingResources{},
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
