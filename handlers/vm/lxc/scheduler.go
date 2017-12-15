package lxc

import (
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type LxcScheduler struct {
}

func NewScheduler() *LxcScheduler {
	return new(LxcScheduler)
}

func (*LxcScheduler) ScheduleInstance(ir model.InstanceResource, offers map[string]*mesos.Offer) (bool, error) {
	cpus := ir.GetVcpu()
	mem := ir.GetMemoryGb()

	for _, offer := range offers {
		offeredCpus := vm.GetOfferScalar(offer, "vcpu")
		if int32(offeredCpus) > cpus {
			offeredMem := vm.GetOfferScalar(offer, "mem")
			if int32(offeredMem) > mem {
				return true, nil
			}
		}
	}
	return false, nil
}
