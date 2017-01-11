package backend

import (
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type TimestampFilter struct {
	Time time.Time
}

func (f *TimestampFilter) OnCreate(v proto.Message) error {
	createdAt, err := ptypes.TimestampProto(f.Time)
	if err != nil {
		return err
	}
	setTimestampField(v, "CreatedAt", createdAt)
	return nil
}

func (f *TimestampFilter) OnUpdate(v proto.Message) error {
	updatedAt, err := ptypes.TimestampProto(f.Time)
	if err != nil {
		return err
	}
	setTimestampField(v, "UpdatedAt", updatedAt)
	return nil
}

func setTimestampField(pb proto.Message, field string, tnow *timestamp.Timestamp) error {
	Filter(pb, "CreatedAt", func(v reflect.Value) {
		if !v.IsNil() {
			return
		}
		if !(v.Kind() == reflect.Ptr) {
			return
		}
		_, ok := v.Interface().(*timestamp.Timestamp)
		if ok {
			v.Set(reflect.ValueOf(tnow))
		}
	})
	return nil
}
