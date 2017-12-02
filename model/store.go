package model

import (
	"fmt"

	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
)

// list of slave resource informations
//MEMO what is the requruied parameter to shedule
var storedOffers map[string]*mesos.Offer = make(map[string]*mesos.Offer)

func IsThereSatisfidCreateReq(i *Instance) (bool, error) {
	var instanceResource InstanceResource
	if instanceResource, ok := i.ResourceTemplate().(InstanceResource); !ok {
		//FIXME instanceResource declared and not used
		return false, fmt.Errorf("Templete do not have vcpu and mem", instanceResource)
	}

	ok, err := i.ResourceTemplate().GetScheduledResource().Schedule(instanceResource, storedOffers)
	if err != nil {
		return false, err
	}

	if ok {
		return true, nil
	}
	return false, nil
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
