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

var (
	mesosOffer *mesos.Offer
)

func init() {
	mesosOffer = &mesos.Offer{
		Id: &mesos.OfferID{
			Value:            sP("d39c0128-4822-49a0-9fab-640fba518d53-O590"),
			XXX_unrecognized: []byte{},
		},
		FrameworkId: &mesos.FrameworkID{
			Value:            sP("d39c0128-4822-49a0-9fab-640fba518d53-0004"),
			XXX_unrecognized: []byte{},
		},
		SlaveId: &mesos.SlaveID{
			Value:            sP("d39c0128-4822-49a0-9fab-640fba518d53-S0"),
			XXX_unrecognized: []byte{},
		},
		Hostname: sP("10.141.141.10"),
		Resources: []*mesos.Resource{
			&mesos.Resource{
				Name: sP("cpus"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value:            fP(2),
					XXX_unrecognized: []byte{},
				},
				Ranges:           nil,
				Set:              nil,
				Role:             sP("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name: sP("mem"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value:            fP(1000.0),
					XXX_unrecognized: []byte{},
				},
				Ranges:           nil,
				Set:              nil,
				Role:             sP("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name: sP("disk"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value: fP(34068),
				},
				// XXX_unrecognized: []byte{},
				Ranges:           nil,
				Set:              nil,
				Role:             sP("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name:   sP("ports"),
				Type:   mesos.Value_Type.Enum(mesos.Value_RANGES),
				Scalar: nil,
				Ranges: &mesos.Value_Ranges{
					Range: []*mesos.Value_Range{
						&mesos.Value_Range{
							Begin:            uP(31000),
							End:              uP(32000),
							XXX_unrecognized: []byte{},
						},
					},
					XXX_unrecognized: []byte{},
				},
				Set:              nil,
				Role:             sP("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
		},
		ExecutorIds: []*mesos.ExecutorID{},
		Attributes:  []*mesos.Attribute{},
		Url: &mesos.URL{
			Scheme: sP("http"),
			Address: &mesos.Address{
				Hostname:         sP("10.141.141.10"),
				Ip:               sP("127.0.1.1"),
				Port:             iP(5051),
				XXX_unrecognized: []byte{},
			},
			Path:             sP("/slave(1)"),
			Query:            []*mesos.Parameter{},
			Fragment:         nil,
			XXX_unrecognized: []byte{},
		},
		Unavailability:   nil,
		XXX_unrecognized: []byte{},
	}
}

func TestConvertToOpenVDCOffer(t *testing.T) {
	assert := assert.New(t)

	vOffer := &model.VDCOffer{
		SlaveID: "d39c0128-4822-49a0-9fab-640fba518d53-S0",
		Resources: []model.VDCOfferResource{
			model.VDCOfferResource{
				Name:   "cpus",
				Type:   model.VDCOfferValueScalar,
				Scalar: 2.0,
			},
			model.VDCOfferResource{
				Name:   "mem",
				Type:   model.VDCOfferValueScalar,
				Scalar: 1000.0,
			},
			model.VDCOfferResource{
				Name:   "disk",
				Type:   model.VDCOfferValueScalar,
				Scalar: 34068.0,
			},
			model.VDCOfferResource{
				Name: "ports",
				Type: model.VDCOfferValueRanges,
				Ranges: []model.VDCOfferValueRange{
					model.VDCOfferValueRange{
						Begin: 31000,
						End:   32000,
					},
				},
			},
		},
	}
	m2vOffer := convertToOpenVDCOffer(mesosOffer)
	assert.Equal(m2vOffer, *vOffer)
}
