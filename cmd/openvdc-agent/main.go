package main

import (
	"fmt"
	"os"
	"time"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/spf13/viper"
)

type VDCAgent struct {
	collector resources.ResourceCollector
	resources *model.ComputingResources
	AgentId   string
}

var (
	DefaultConfPath string
	updateInteval time.Duration = 5
	// TODO: receive from configuration file or as flag
	tmpCollectorType = "local"
)

func initConfig() error {
	viper.SetConfigFile(DefaultConfPath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if viper.ConfigFileUsed() == DefaultConfPath && os.IsNotExist(err) {
			// Ignore default conf file does not exist.
			return nil
		}
		return err
	}
	return nil
}

func main() {
	exitOnErr := func(err error) {
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
	exitOnErr(initConfig())
	exitOnErr(newVDCAgent(exitOnErr).Run())
}

func newVDCAgent(exitCallback func(err error)) *VDCAgent {
	c, err := resources.NewCollector(tmpCollectorType, viper.GetViper());
	exitCallback(err)

	return &VDCAgent{
		collector: c,
		resources: &model.ComputingResources{},
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
