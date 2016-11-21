package model

import (
	"fmt"
	"path"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
)

type InstanceOps interface {
	Create(*Instance) (*Instance, error)
	FindByID(string) (*Instance, error)
}

const instancesBaseKey = "instances"

func init() {
	schemaKeys = append(schemaKeys, instancesBaseKey)
}

type instances struct {
	ctx context.Context
}

func Instances(ctx context.Context) InstanceOps {
	return &instances{ctx: ctx}
}

func (i *instances) connection() (backend.ModelBackend, error) {
	bk := GetBackendCtx(i.ctx)
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	return bk, nil
}

func (i *instances) Create(n *Instance) (*Instance, error) {
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID("/instances/i-", data)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	return n, nil
}

func (i *instances) FindByID(id string) (*Instance, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	v, err := bk.Find(fmt.Sprintf("/instances/%s", id))
	if err != nil {
		return nil, err
	}
	n := &Instance{}
	err = proto.Unmarshal(v, n)
	if err != nil {
		return nil, err
	}
	n.Id = id
	return n, nil
}
