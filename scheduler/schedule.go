package scheduler

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

var (
	instanceSchedulerHandlers = make(map[string]InstanceScheduleHandler)
)

func RegisterInstanceScheduleHandler(name string, i InstanceScheduleHandler) error {
	if _, exists := instanceSchedulerHandlers[name]; exists {
		return fmt.Errorf("Duplicated name for instance schedule handler: %s", name)
	}
	instanceSchedulerHandlers[name] = i
	return nil
}

type Schedule struct {
	sync.Mutex
	storedOffers map[string]*model.VDCOffer
}

// openvdc offer (interface | struct) の定義をここに書く
func (s *Schedule) StoreOffer(offer *model.VDCOffer) {
	s.Lock()
	defer s.Unlock()
	storedOffers[offer.SlaveID] = offer
}

func (s *Schedule) Assign(inst *model.Instance) error {
	flog := log.WithFields(log.Fields{
		"instance_id": in.GetInstanceId(),
	})

	name := inst.ResourceTemplate().ResourceName
	instSchedHandler := instanceSchedulerHandlers[name]
	for _, offer := range s.storedOffers {
		ok, err := instSchedHandler.ScheduleInstance(inst, offer)
		if err != nil {
			return err
		}
		if ok {
			flog.Infof("Assined")
			return nil
		}
	}
	return fmt.Errorf("There is no machine can satisfy resource requirement")
}

// Generic scheduling util functions
// copy from https://github.com/mesosphere/mesos-framework-tutorial/blob/master/scheduler/utils.go
func (s *Schedule) GetOfferScalar(offer *model.VDCOffer, name string) float64 {
	resources := filterResources(offer.Resources, func(res model.Resource) bool {
		return res.Name == name
	})

	value := 0.0
	for _, res := range resources {
		value += res.Scalar
	}
	return value
}

func filterResources(resources []model.Resource, filter func(model.Resource) bool) (result []model.Resource) {
	for _, resource := range resources {
		if filter(resource) {
			result = append(result, resource)
		}
	}
	return result
}

func (s *Schedule) IsThereSatisfidCreateReq(i *model.Instance) (bool, error) {
	var instanceResource model.InstanceResource
	var instResHandler handlers.InstanceResourceHandler
	if instResHandler, ok := i.is.(*model.InstanceResource); !ok {
		//FIXME instanceResource d eclared and not used
		return false, fmt.Errorf("Templete do not have vcpu and mem", instanceResource)
	}

	ok, err := instResHandler.GetInstanceSchedulerHandler().ScheduleInstance(instanceResource, storedOffers)
	if err != nil {
		return false, err
	}
	return ok, nil
}
