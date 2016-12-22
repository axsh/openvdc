package model

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func FilterCreatedAt(pb proto.Message, tnow *timestamp.Timestamp) proto.Message {
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
	return pb
}

func Filter(pb proto.Message, field string, cb func(reflect.Value)) {
	v := reflect.ValueOf(pb)
	if !v.IsValid() || v.IsNil() {
		return
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Type().Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			if v.Type().Field(i).Name == field {
				cb(fv)
			}
			fpb, ok := fv.Interface().(proto.Message)
			if ok {
				Filter(fpb, field, cb)
			}
		}
	}
}
