package main

import (
	"fmt"

	"github.com/axsh/openvdc/model"
)

type esxiResourceCollector struct {
	hostIp   string
	hostUser string
	hostPwd  string
}

func initConfig() {
	fmt.Println("init esxi config")
}

func NewEsxiResourceCollector() (ResourceCollector, error) {
	initConfig()
	return &esxiResourceCollector{}, nil
}

func (rm *esxiResourceCollector) GetCpu() (*model.Resource, error) {
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetMem() (*model.Resource,error) {
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetDisk() ([]*model.Resource, error) {
	disks := make([]*model.Resource, 0)
	return disks, nil
}

func (rm *esxiResourceCollector) GetLoadAvg() (*model.LoadAvg, error) {
	return &model.LoadAvg{}, nil
}
