package null

import (
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/scheduler"
)

func init() {
	scheduler.RegisterInstanceScheduleHandler("vm/null", &NullScheduler{})
}

type NullScheduler struct {
}

func NewScheduler() *NullScheduler {
	return new(NullScheduler)
}

func (*NullScheduler) ScheduleInstance(ir model.InstanceResource, offer model.VDCOffer) (bool, error) {
	return true, nil
}
