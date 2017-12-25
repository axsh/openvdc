package lxc

import (
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
)

type LxcScheduler struct {
}

func NewScheduler() *LxcScheduler {
	return new(LxcScheduler)
}

func (*LxcScheduler) ScheduleInstance(ir model.InstanceResource, offer model.VDCOffer) (bool, error) {
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
