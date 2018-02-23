package scheduler

import (
	"testing"

	"github.com/axsh/openvdc/model"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/stretchr/testify/assert"
)

var (
	mesosOffer *mesos.Offer
)

func init() {
	mesosOffer = &mesos.Offer{
		Id: &mesos.OfferID{
			Value:            proto.String("d39c0128-4822-49a0-9fab-640fba518d53-O590"),
			XXX_unrecognized: []byte{},
		},
		FrameworkId: &mesos.FrameworkID{
			Value:            proto.String("d39c0128-4822-49a0-9fab-640fba518d53-0004"),
			XXX_unrecognized: []byte{},
		},
		SlaveId: &mesos.SlaveID{
			Value:            proto.String("d39c0128-4822-49a0-9fab-640fba518d53-S0"),
			XXX_unrecognized: []byte{},
		},
		Hostname: proto.String("10.141.141.10"),
		Resources: []*mesos.Resource{
			&mesos.Resource{
				Name: proto.String("cpus"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value:            proto.Float64(2),
					XXX_unrecognized: []byte{},
				},
				Ranges:           nil,
				Set:              nil,
				Role:             proto.String("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name: proto.String("mem"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value:            proto.Float64(1000.0),
					XXX_unrecognized: []byte{},
				},
				Ranges:           nil,
				Set:              nil,
				Role:             proto.String("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name: proto.String("disk"),
				Type: mesos.Value_Type.Enum(mesos.Value_SCALAR),
				Scalar: &mesos.Value_Scalar{
					Value: proto.Float64(34068),
				},
				// XXX_unrecognized: []byte{},
				Ranges:           nil,
				Set:              nil,
				Role:             proto.String("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
			&mesos.Resource{
				Name:   proto.String("ports"),
				Type:   mesos.Value_Type.Enum(mesos.Value_RANGES),
				Scalar: nil,
				Ranges: &mesos.Value_Ranges{
					Range: []*mesos.Value_Range{
						&mesos.Value_Range{
							Begin:            proto.Uint64(31000),
							End:              proto.Uint64(32000),
							XXX_unrecognized: []byte{},
						},
					},
					XXX_unrecognized: []byte{},
				},
				Set:              nil,
				Role:             proto.String("*"),
				Disk:             nil,
				Reservation:      nil,
				Revocable:        nil,
				XXX_unrecognized: []byte{},
			},
		},
		ExecutorIds: []*mesos.ExecutorID{},
		Attributes:  []*mesos.Attribute{},
		Url: &mesos.URL{
			Scheme: proto.String("http"),
			Address: &mesos.Address{
				Hostname:         proto.String("10.141.141.10"),
				Ip:               proto.String("127.0.1.1"),
				Port:             proto.Int32(5051),
				XXX_unrecognized: []byte{},
			},
			Path:             proto.String("/slave(1)"),
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
	assert.Equal(*m2vOffer, *vOffer)
}
