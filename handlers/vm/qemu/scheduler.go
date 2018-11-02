package qemu

import (
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/scheduler"
)

func init() {
	scheduler.RegisterInstanceScheduleHandler("vm/qemu", &QemuScheduler{})
}

type QemuScheduler struct {
}

func NewScheduler() *QemuScheduler {
	return new(QemuScheduler)
}

func (*QemuScheduler) ScheduleInstance(ir model.InstanceResource, offer *model.VDCOffer) (bool, error) {
	cpus := ir.GetVcpu()
	mem := ir.GetMemoryGb()

	offerCpus := int32(vm.GetOfferScalar(offer, "cpus"))
	offerMem := int32(vm.GetOfferScalar(offer, "mem") / 1000)
	if cpus < offerCpus && mem < offerMem {
		return true, nil
	}
	return false, nil
}
