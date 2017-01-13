package backend

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// ProtoModelBackend is the ModelBackend interface wrapper
// that accepts proto.Message instead of raw bytes.
type ProtoModelBackend interface {
	Backend() ModelBackend
	AddFilter(f ProtoFilter)
	Create(key string, value proto.Message) error
	CreateWithID(key string, value proto.Message) (string, error)
	Update(key string, value proto.Message) error
	Find(key string, v proto.Message) error
	Delete(key string) error
	Keys(parentKey string) (KeyIterator, error)
	FindLastKey(prefixKey string) (string, error)
}

type ProtoFilter interface {
	OnCreate(v proto.Message) error
	OnUpdate(v proto.Message) error
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

type ProtoWrapper struct {
	backend ModelBackend
	filters []ProtoFilter
}

func NewProtoWrapper(bk ModelBackend) *ProtoWrapper {
	return &ProtoWrapper{bk, []ProtoFilter{}}
}

func (p *ProtoWrapper) Backend() ModelBackend {
	return p.backend
}

func (p *ProtoWrapper) AddFilter(f ProtoFilter) {
	p.filters = append(p.filters, f)
}

func (p *ProtoWrapper) Create(key string, value proto.Message) error {
	for _, f := range p.filters {
		err := f.OnCreate(value)
		if err != nil {
			return errors.Wrapf(err, "Failed to apply %T.OnCreate", f)
		}
	}

	buf, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return p.backend.Create(key, buf)
}

func (p *ProtoWrapper) CreateWithID(key string, value proto.Message) (string, error) {
	for _, f := range p.filters {
		err := f.OnCreate(value)
		if err != nil {
			return "", errors.Wrapf(err, "Failed to apply %T.OnCreate", f)
		}
	}
	buf, err := proto.Marshal(value)
	if err != nil {
		return "", err
	}

	return p.backend.CreateWithID(key, buf)
}

func (p *ProtoWrapper) Update(key string, value proto.Message) error {
	for _, f := range p.filters {
		err := f.OnUpdate(value)
		if err != nil {
			return errors.Wrapf(err, "Failed to apply %T.OnUpdate", f)
		}
	}
	buf, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return p.backend.Update(key, buf)
}

func (p *ProtoWrapper) Find(key string, v proto.Message) error {
	buf, err := p.backend.Find(key)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, v)
}

func (p *ProtoWrapper) Delete(key string) error {
	return p.backend.Delete(key)
}

func (p *ProtoWrapper) Keys(parentKey string) (KeyIterator, error) {
	return p.backend.Keys(parentKey)
}

func (p *ProtoWrapper) FindLastKey(prefixKey string) (string, error) {
	return p.backend.FindLastKey(prefixKey)
}
