package scheduler

import (
	"testing"

	"github.com/axsh/openvdc/model"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/stretchr/testify/assert"
)

// to pointer
func sP(s string) *string {
	return &s
}

func fP(s float64) *float64 {
	return &s
}

func iP(s int) *int32 {
	i32 := int32(s)
	return &i32
}

func uP(s uint64) *uint64 {
	return &s
}

func TestConvertToOpenVDCOffer(t *testing.T) {
	assert := assert.New(t)
	mOffer := &mesos.Offer{
		Id: &mesos.OfferID{
			Value: sP("d39c0128-4822-49a0-9fab-640fba518d53-O590"),
			// XXX_unrecognized:[],
		},
		FrameworkId: &mesos.FrameworkID{
			Value: sP("d39c0128-4822-49a0-9fab-640fba518d53-0004"),
			// XXX_unrecognized:[],
		},
		SlaveId: &mesos.SlaveID{
			Value: sP("d39c0128-4822-49a0-9fab-640fba518d53-S0"),
			// XXX_unrecognized:[],
		},
		Hostname: sP("10.141.141.10"),
		Resources: []*mesos.Resource{
			&mesos.Resource{
				Name: sP("cpus"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value: fP(2),
					// XXX_unrecognized:[],
				},
				Ranges: nil,
				Set:    nil,
				// Role:**,???
				Disk:        nil,
				Reservation: nil,
				Revocable:   nil,
				// XXX_unrecognized:[],
			},
			&mesos.Resource{
				Name: sP("mem"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value: fP(1000.0),
					// XXX_unrecognized:[],
				},
				Ranges: nil,
				Set:    nil,
				// Role:**,???
				Disk:        nil,
				Reservation: nil,
				Revocable:   nil,
				//XXX_unrecognized:[],
			},
			&mesos.Resource{
				Name: sP("disk"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value: fP(34068),
				},
				// XXX_unrecognized:[],

				Ranges: nil,
				Set:    nil,
				// Role:**,
				Disk:        nil,
				Reservation: nil,
				Revocable:   nil,
				//XXX_unrecognized:[],
			},
			&mesos.Resource{
				Name:   sP("ports"),
				Type:   mesos.Value_Type.Enum(mesos.Value_RANGES),
				Scalar: nil,
				Ranges: &mesos.Value_Ranges{
					Range: []*mesos.Value_Range{
						&mesos.Value_Range{
							Begin: uP(31000),
							End:   uP(32000),
							//XXX_unrecognized:[],
						},
					},
					// XXX_unrecognized:[],
				},
				Set: nil,
				//Role:**,
				Disk:        nil,
				Reservation: nil,
				Revocable:   nil,
				// XXX_unrecognized:[],
			},
		},
		//ExecutorIds:[],
		//Attributes:[],
		Url: &mesos.URL{
			Scheme: sP("http"),
			Address: &mesos.Address{
				Hostname: sP("10.141.141.10"),
				Ip:       sP("127.0.1.1"),
				Port:     iP(5051),
				//XXX_unrecognized:[],
			},
			//Path:*/slave(1),
			//Query:[],
			Fragment: nil,
			//XXX_unrecognized:[],
		},
		Unavailability: nil,
		//XXX_unrecognized:[],
	}

	vOffer := &model.VDCOffer{
		SlaveID: "test",
		Resources: []model.Resource{
			model.Resource{
				Name:   "vcpu",
				Type:   model.ValueScalar,
				Scalar: 1.0,
			},
		},
	}
	m2vOffer := convertToOpenVDCOffer(mOffer)
	assert.Equal(m2vOffer, vOffer)
}
