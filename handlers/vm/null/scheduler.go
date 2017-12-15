package null

import (
	"github.com/axsh/openvdc/model"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type NullScheduler struct {
}

func NewScheduler() *NullScheduler {
	return new(NullScheduler)
}

func (*NullScheduler) ScheduleInstance(ir model.InstanceResource, offers map[string]*mesos.Offer) (bool, error) {
	return true, nil
}
