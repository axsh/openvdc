package model

import (
	"errors"
	"fmt"
	"path"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/proto"
)

var ErrInstanceMissingResource = errors.New("Resource is not associated")

type InstanceOps interface {
	Create(*Instance) (*Instance, error)
	FindByID(string) (*Instance, error)
	UpdateState(id string, next Instance_State) error
	FilterByState(state Instance_State) ([]*Instance, error)
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
	if n.GetResourceId() == "" {
		return nil, ErrInstanceMissingResource
	}
	n.State = Instance_REGISTERED
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID(fmt.Sprintf("/%s/i-", instancesBaseKey), data)
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
	v, err := bk.Find(fmt.Sprintf("/%s/%s", instancesBaseKey, id))
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

func (i *instances) UpdateState(id string, next Instance_State) error {
	n, err := i.FindByID(id)
	if err != nil {
		return err
	}
	if err := n.validateStateTransition(next); err != nil {
		return err
	}
	n.State = next

	bk, err := i.connection()
	if err != nil {
		return err
	}
	buf, err := proto.Marshal(n)
	if err != nil {
		return err
	}
	return bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, id), buf)
}

func (i *instances) FilterByState(state Instance_State) ([]*Instance, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	res := []*Instance{}
	keys, err := bk.Keys(fmt.Sprintf("/%s", instancesBaseKey))
	if err != nil {
		return nil, err
	}
	for keys.Next() {
		instance, err := i.FindByID(keys.Value())
		if err != nil {
			return nil, err
		}
		if instance.GetState() == state {
			res = append(res, instance)
		}
	}
	return res, nil
}

func (i *Instance) Resource(ctx context.Context) (*Resource, error) {
	return Resources(ctx).FindByID(i.GetResourceId())
}

func (i *Instance) validateStateTransition(next Instance_State) error {
	result := func() bool {
		switch i.GetState() {
		case Instance_REGISTERED:
			return (next == Instance_QUEUED)
		case Instance_QUEUED:
			return (next == Instance_STARTING)
		case Instance_STARTING:
			return (next == Instance_RUNNING)
		case Instance_RUNNING:
			return (next == Instance_STOPPING ||
				next == Instance_SHUTTINGDOWN)
		case Instance_STOPPING:
			return (next == Instance_STOPPED)
		case Instance_STOPPED:
			return (next == Instance_STARTING ||
				next == Instance_SHUTTINGDOWN)
		case Instance_SHUTTINGDOWN:
			return (next == Instance_TERMINATED)
		}
		return false
	}()

	if result {
		return nil
	}

	return fmt.Errorf("Invalid state transition: %s -> %s",
		i.GetState().String(),
		next.String(),
	)
}
