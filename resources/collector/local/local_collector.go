package local

import (
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/spf13/viper"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type localResourceCollector struct{}

func init() {
	resources.RegisterCollector("local", NewLocalResourceCollector)
}


func NewLocalResourceCollector(conf *viper.Viper) (resources.ResourceCollector, error) {
	return &localResourceCollector{}, nil
}

func (rc *localResourceCollector) GetCpu() (*model.Resource, error) {
	cpu, _ := cpu.Info()
	return &model.Resource{
		Total: int64(len(cpu)),
	}, nil
}

func (rc *localResourceCollector) GetMem() (*model.Resource, error) {
	mem, _ := mem.VirtualMemory()
	return &model.Resource{
		Total:       int64(mem.Total),
		Available:   int64(mem.Available),
		UsedPercent: mem.UsedPercent,
	}, nil
}

func (rc *localResourceCollector) GetDisk() ([]*model.Resource, error) {
	dp, _ := disk.Partitions(true)
	disks := make([]*model.Resource, 0)

	for _, part := range dp {
		d, _ := disk.Usage(part.Mountpoint)
		disks = append(disks, &model.Resource{
			Total:       int64(d.Total),
			Available:   int64(d.Free),
			UsedPercent: d.UsedPercent,
		})
	}
	return disks, nil
}

func (rc *localResourceCollector) GetLoadAvg() (*model.LoadAvg, error) {
	l, _ := load.Avg()
	return &model.LoadAvg{
		Load1:  float32(l.Load1),
		Load5:  float32(l.Load15),
		Load15: float32(l.Load15),
	}, nil
}
