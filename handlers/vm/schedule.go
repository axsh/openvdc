package vm

import (
	"fmt"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
)

type Schedule struct {
	// list of slave resource informations
	//MEMO what is the requruied parameter to shedule
	storedOffers map[string]*mesos.Offer
}

// openvdc offer (interface | struct) の定義をここに書く
func (s *Schedule) StoreOffer(offer *mesos.Offer) {
	if s.storedOffers == nil {
		s.storedOffers = make(map[string]*mesos.Offer)
	}
	s.storedOffers[offer.SlaveId.GetValue()] = offer
}

// Generic scheduling util functions
// copy from https://github.com/mesosphere/mesos-framework-tutorial/blob/master/scheduler/utils.go
func (s *Schedule) GetOfferScalar(offer *mesos.Offer, name string) float64 {
	resources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
		return res.GetName() == name
	})

	value := 0.0
	for _, res := range resources {
		value += res.GetScalar().GetValue()
	}
	return value
}

func (s *Schedule) IsThereSatisfidCreateReq(i *model.Instance) (bool, error) {
	var instanceResource *model.InstanceResource
	var instResHandler *handlers.InstanceResourceHandler
	if instResHandler, ok := i.ResourceTemplate().(*model.InstanceResource); !ok {
		//FIXME instanceResource declared and not used
		return false, fmt.Errorf("Templete do not have vcpu and mem", instanceResource)
	}

	ok, err := instResHandler.GetInstanceSchedulerHandler().Schedule(instanceResource, s.storedOffers)
	if err != nil {
		return false, err
	}
	return ok, nil
}
