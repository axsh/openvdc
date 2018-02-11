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
	schedule                  *Schedule
)

func init() {
	schedule = newSchedule()
}

func RegisterInstanceScheduleHandler(name string, i handlers.InstanceScheduleHandler) error {
	if _, exists := instanceSchedulerHandlers[name]; exists {
		return fmt.Errorf("Duplicated name for instance schedule handler: %s", name)
	}
	instanceSchedulerHandlers[name] = i
	return nil
}

type Schedule struct {
	*sync.Mutex
	storedOffers map[string]model.VDCOffer
}

func newSchedule() *Schedule {
	return &Schedule{
		Mutex:        new(sync.Mutex),
		storedOffers: make(map[string]model.VDCOffer),
	}
}

// openvdc offer (interface | struct) の定義をここに書く
func (s *Schedule) StoreOffer(offer model.VDCOffer) {
	schedule.Lock()
	defer schedule.Unlock()
	schedule.storedOffers[offer.SlaveID] = offer
}

func (s *Schedule) Assign(inst *model.Instance) error {
	flog := log.WithFields(log.Fields{
		"instance_id": inst.GetId(), // ASK GetInstanceId() -> GetId()
	})

	name := inst.ResourceTemplate().ResourceName()
	var instSchedHandler handlers.InstanceScheduleHandler
	var ok bool
	if instSchedHandler, ok = instanceSchedulerHandlers[name]; !ok {
		return fmt.Errorf("%s instanceSchedulerHandlers is not registered", name)
	}
	var instResource model.InstanceResource
	instResource, ok = inst.ResourceTemplate().(model.InstanceResource)
	if !ok {
		return fmt.Errorf("Templete do not have vcpu and mem", instResource)
	}
	for _, offer := range schedule.storedOffers {
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
