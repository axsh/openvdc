package model

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
)

// list of slave resource informations
var storedOffers map[string]*mesos.Offer = make(map[string]*mesos.Offer)

func IsThereSatisfidCreateReq(i *Instance) (bool, error) {

	var instanceResource InstanceResource
	if instanceResource, ok := i.ResourceTemplate().(InstanceResource); !ok {
		//FIXME instanceResource declared and not used
		return false, fmt.Errorf("Templete do not have vcpu and mem", instanceResource)
	}

	// TODO: Avoid type switch to find template types.
	var offerParams []string
	var reqValues []int32
	switch t := i.GetTemplate().GetItem(); t.(type) {
	case *Template_Lxc:
		offerParams = []string{"vcpu", "mem"}
		reqValues = []int32{
			instanceResource.GetVcpu(),
			instanceResource.GetMemoryGb(),
		}
	case *Template_Null:

	case *Template_Qemu:
		offerParams = []string{"vcpu", "mem"}
		reqValues = []int32{
			instanceResource.GetVcpu(),
			instanceResource.GetMemoryGb(),
		}
	case *Template_Esxi:

	default:
		log.Warnf("Unknown template type: %T", t)
		return true, fmt.Errorf("Unknown template type: %T", t)
	}

	for _, offer := range storedOffers {
		var index int
		for i, param := range offerParams {
			index = i
			offerValue := int32(getOfferScalar(offer, param))
			if offerValue < reqValues[i] {
				break
			}
		}
		if index == len(offerParams)-1 {
			return true, nil
		}
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
