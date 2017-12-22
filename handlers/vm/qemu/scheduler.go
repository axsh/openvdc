package qemu

import (
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
)

func init() {
	scheduler.RegisterInstanceScheduleHandler(&QemuScheduler{})
}

type QemuScheduler struct {
}

func NewScheduler() *QemuScheduler {
	return new(QemuScheduler)
}

func (*QemuScheduler) ScheduleInstance(ir model.InstanceResource, offer *model.VDCOffer) (bool, error) {
	cpus := ir.GetVcpu()
	mem := ir.GetMemoryGb()

	offeredCpus := vm.GetOfferScalar(offer, "vcpu")
	if int32(offeredCpus) > cpus {
		offeredMem := vm.GetOfferScalar(offer, "mem")
		if int32(offeredMem) > mem {
			return true, nil
		}
	}
	return false, nil
}
