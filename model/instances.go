package model

import (
	"errors"
	"fmt"
	"path"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
)

var ErrInstanceMissingResource = errors.New("Resource is not associated")

type InstanceOps interface {
	Create(*Instance) (*Instance, error)
	FindByID(string) (*Instance, error)
	UpdateState(id string, next InstanceState) error
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
	n.State = InstanceState_INSTANCE_REGISTERED
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

func (i *instances) UpdateState(id string, next InstanceState) error {
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

func (i *Instance) Resource(ctx context.Context) (*Resource, error) {
	return Resources(ctx).FindByID(i.GetResourceId())
}

func (i *Instance) validateStateTransition(next InstanceState) error {
	result := func() bool {
		switch i.GetState() {
		case InstanceState_INSTANCE_REGISTERED:
			return (next == InstanceState_INSTANCE_QUEUED)
		case InstanceState_INSTANCE_QUEUED:
			return (next == InstanceState_INSTANCE_STARTING)
		case InstanceState_INSTANCE_STARTING:
			return (next == InstanceState_INSTANCE_RUNNING)
		case InstanceState_INSTANCE_RUNNING:
			return (next == InstanceState_INSTANCE_STOPPING ||
				next == InstanceState_INSTANCE_SHUTTINGDOWN)
		case InstanceState_INSTANCE_STOPPING:
			return (next == InstanceState_INSTANCE_STOPPED)
		case InstanceState_INSTANCE_STOPPED:
			return (next == InstanceState_INSTANCE_STARTING ||
				next == InstanceState_INSTANCE_SHUTTINGDOWN)
		case InstanceState_INSTANCE_SHUTTINGDOWN:
			return (next == InstanceState_INSTANCE_TERMINATED)
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
