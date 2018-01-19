package scheduler

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

var (
	instanceSchedulerHandlers = make(map[string]handlers.InstanceScheduleHandler)
)

func RegisterInstanceScheduleHandler(name string, i handlers.InstanceScheduleHandler) error {
	if _, exists := instanceSchedulerHandlers[name]; exists {
		return fmt.Errorf("Duplicated name for instance schedule handler: %s", name)
	}
	instanceSchedulerHandlers[name] = i
	return nil
}

type Schedule struct {
	sync.Mutex
	storedOffers map[string]model.VDCOffer
}

// openvdc offer (interface | struct) の定義をここに書く
func (s *Schedule) StoreOffer(offer model.VDCOffer) {
	s.Lock()
	defer s.Unlock()
	s.storedOffers[offer.SlaveID] = offer
}

func (s *Schedule) Assign(inst *model.Instance) error {
	flog := log.WithFields(log.Fields{
		"instance_id": inst.GetId(), // ASK GetInstanceId() -> GetId()
	})

	name := inst.ResourceTemplate().ResourceName()
	instSchedHandler := instanceSchedulerHandlers[name]
	var instResource model.InstanceResource
	instResource, ok := inst.ResourceTemplate().(model.InstanceResource)
	if !ok {
		return fmt.Errorf("Templete do not have vcpu and mem", instResource)
	}
	for _, offer := range s.storedOffers {
		ok, err := instSchedHandler.ScheduleInstance(instResource, offer)
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
