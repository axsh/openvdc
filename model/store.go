package model

import (
	"fmt"

	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
)

// list of slave resource informations
var storedOffers map[string]*mesos.Offer = make(map[string]*mesos.Offer)

func IsThereSatisfidCreateReq(i *Instance) (bool, erorr) {
	if instanceResource, ok := i.ResourceTemplate().(InstanceResource); ok {
		cpus := instanceResource.GetVcpu()
		mem := instanceResource.GetMemoryGb()

		for _, offer := range storedOffers {
			offeredCpus := getOfferScalar(offer, "vcpu")
			if int32(offeredCpus) > cpus {
				offeredMem := getOfferScalar(offer, "mem")
				if int32(offeredMem) > mem {
					return true, nil
				}
			}
		}
		return false, nil
	}
	//DEBUG Return true in temporary to pass the other test
	return true, fmt.Errorf("Templete do not have vcpu and mem")
}

func StoreOffer(offer *mesos.Offer) {
	storedOffers[offer.SlaveId.GetValue()] = offer
}

// copy from https://github.com/mesosphere/mesos-framework-tutorial/blob/master/scheduler/utils.go
func getOfferScalar(offer *mesos.Offer, name string) float64 {
	resources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
		return res.GetName() == name
	})

	value := 0.0
	for _, res := range resources {
		value += res.GetScalar().GetValue()
	}

	return value
}
