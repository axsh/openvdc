package qemu

import (
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/scheduler"
)

func init() {
	scheduler.RegisterInstanceScheduleHandler("qemu", &QemuScheduler{})
}

type QemuScheduler struct {
}

func NewScheduler() *QemuScheduler {
	return new(QemuScheduler)
}

func (q *QemuScheduler) ScheduleInstance(ir model.InstanceResource, offer model.VDCOffer) (bool, error) {
	cpus := ir.GetVcpu()
	mem := ir.GetMemoryGb()

	offerCpus := int32(vm.GetOfferScalar(offer, "vcpu"))
	offerMem := int32(vm.GetOfferScalar(offer, "mem"))
	if offerCpus < cpus || offerMem < mem {
		return false, nil
	}
	return true, nil
}
