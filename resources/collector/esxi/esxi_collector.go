package esxi

import (
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/spf13/viper"
)

type esxiResourceCollector struct {
	hostIp       string
	hostUser     string
	hostPwd      string
	hostInsecure bool
}

func init() {
	resources.RegisterCollector("esxi", NewEsxiResourceCollector)
}

func NewEsxiResourceCollector(conf *viper.Viper) (resources.ResourceCollector, error) {
	viper.SetDefault("hypervisor.esxi-insecure", true)

	return &esxiResourceCollector{
		hostIp:       conf.GetString("hypervisor.esxi-ip"),
		hostUser:     conf.GetString("hypervisor.esxi-user"),
		hostPwd:      conf.GetString("hypervisor.esxi-pass"),
		hostInsecure: conf.GetBool("hypervisor.esxi-insecure"),
	}, nil
}

func (rm *esxiResourceCollector) GetCpu() (*model.Resource, error) {
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetMem() (*model.Resource, error) {
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetDisk() ([]*model.Resource, error) {
	disks := make([]*model.Resource, 0)
	return disks, nil
}

func (rm *esxiResourceCollector) GetLoadAvg() (*model.LoadAvg, error) {
	return &model.LoadAvg{}, nil
}
