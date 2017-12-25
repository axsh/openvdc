package null

import (
	"github.com/axsh/openvdc/model"
)

type NullScheduler struct {
}

func NewScheduler() *NullScheduler {
	return new(NullScheduler)
}

func (*NullScheduler) ScheduleInstance(ir model.InstanceResource, offer model.VDCOffer) (bool, error) {
	return true, nil
}
