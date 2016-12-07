package model

import (
	"errors"
	"fmt"
	"path"
	"time"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

var ErrInstanceMissingResource = errors.New("Resource is not associated")
var ErrInvalidID = errors.New("ID is missing")

type InstanceOps interface {
	Create(*Instance) (*Instance, error)
	FindByID(string) (*Instance, error)
	UpdateState(id string, next InstanceState_State) error
	FilterByState(state InstanceState_State) ([]*Instance, error)
	Update(*Instance) error
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

	createdAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}
	initState := &InstanceState{
		State:     InstanceState_REGISTERED,
		CreatedAt: createdAt,
	}
	n.LastState = initState
	data1, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	data2, err := proto.Marshal(initState)
	if err != nil {
		return nil, err
	}

	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID(fmt.Sprintf("/%s/i-", instancesBaseKey), data1)
	if err != nil {
		return nil, err
	}
	_, err = bk.Find(nkey)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	if err = bk.Create(fmt.Sprintf("%s/state", nkey), []byte{}); err != nil {
		return nil, err
	}
	_, err = bk.CreateWithID(fmt.Sprintf("%s/state/state-", nkey), data2)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (i *instances) Update(instance *Instance) error {
	if instance.Id == "" {
		return ErrInvalidID
	}

	buf, err := proto.Marshal(instance)
	if err != nil {
		return err
	}
	bk, err := i.connection()
	if err != nil {
		return err
	}
	err = bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, instance.Id), buf)
	if err != nil {
		return err
	}
	return nil
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

func (i *instances) findLastState(id string) (*InstanceState, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	key, err := bk.FindLastKey(fmt.Sprintf("/%s/%s/state/state-", instancesBaseKey, id))
	if err != nil {
		return nil, err
	}
	buf, err := bk.Find(key)
	if err != nil {
		return nil, err
	}

	n := &InstanceState{}
	err = proto.Unmarshal(buf, n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (i *instances) UpdateState(id string, next InstanceState_State) error {
	instance, err := i.FindByID(id)
	if err != nil {
		return err
	}
	if err := instance.LastState.validateStateTransition(next); err != nil {
		return err
	}
	createdAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	nstate := &InstanceState{
		State:     next,
		CreatedAt: createdAt,
	}
	instance.LastState = nstate

	buf1, err := proto.Marshal(instance)
	if err != nil {
		return err
	}
	buf2, err := proto.Marshal(nstate)
	if err != nil {
		return err
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	_, err = bk.CreateWithID(fmt.Sprintf("/%s/%s/state/state-", instancesBaseKey, id), buf2)
	if err != nil {
		return err
	}
	return bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, id), buf1)
}

func (i *instances) FilterByState(state InstanceState_State) ([]*Instance, error) {
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
		if instance.LastState.State == state {
			res = append(res, instance)
		}
	}
	return res, nil
}

func (i *Instance) Resource(ctx context.Context) (*Resource, error) {
	return Resources(ctx).FindByID(i.GetResourceId())
}

func (i *InstanceState) validateStateTransition(next InstanceState_State) error {
	result := func() bool {
		switch i.State {
		case InstanceState_REGISTERED:
			return (next == InstanceState_QUEUED)
		case InstanceState_QUEUED:
			return (next == InstanceState_STARTING)
		case InstanceState_STARTING:
			return (next == InstanceState_RUNNING)
		case InstanceState_RUNNING:
			return (next == InstanceState_STOPPING ||
				next == InstanceState_SHUTTINGDOWN)
		case InstanceState_STOPPING:
			return (next == InstanceState_STOPPED)
		case InstanceState_STOPPED:
			return (next == InstanceState_STARTING ||
				next == InstanceState_SHUTTINGDOWN)
		case InstanceState_SHUTTINGDOWN:
			return (next == InstanceState_TERMINATED)
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
