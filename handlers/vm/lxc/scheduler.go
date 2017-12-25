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

	offerCpus := int32(vm.GetOfferScalar(offer, "vcpu"))
	offerMem := int32(vm.GetOfferScalar(offer, "mem"))
	if offerCpus < cpus || offerMem < mem {
		return false, nil
	}
	return true, nil
}
