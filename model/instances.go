package model

import (
	"fmt"
	"path"
	"time"

	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var ErrInvalidID = errors.New("ID is missing")

type InstanceOps interface {
	Create(*Instance) (*Instance, error)
	FindByID(string) (*Instance, error)
	AddFailureMessage(id string, failureMessage FailureMessage_ErrorType) error
	//GetLatestFailureMessage(id string) (*FailureMessage, error)
	UpdateState(id string, next InstanceState_State) error
	FilterByState(state InstanceState_State) ([]*Instance, error)
	Update(*Instance) error
	Filter(limit int, cb func(*Instance) int) error
	WaitStateUpdate(id string) (*InstanceState, error)
}

const instancesBaseKey = "instances"

type stateDef struct {
	Nexts InstanceStateSlice
	Goals InstanceStateSlice
}

var instanceStateDefs []*stateDef = make([]*stateDef, len(InstanceState_State_name))

func init() {
	schemaKeys = append(schemaKeys, instancesBaseKey)

	instanceStateDefs[InstanceState_REGISTERED] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_QUEUED, InstanceState_TERMINATED},
		Goals: []InstanceState_State{InstanceState_QUEUED, InstanceState_TERMINATED},
	}
	instanceStateDefs[InstanceState_QUEUED] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_STARTING},
		Goals: []InstanceState_State{InstanceState_RUNNING, InstanceState_STOPPED},
	}
	instanceStateDefs[InstanceState_STARTING] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_RUNNING, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_RUNNING},
	}
	instanceStateDefs[InstanceState_RUNNING] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_STOPPING, InstanceState_SHUTTINGDOWN, InstanceState_REBOOTING, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_STOPPED, InstanceState_TERMINATED, InstanceState_RUNNING},
	}
	instanceStateDefs[InstanceState_STOPPING] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_STOPPED, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_STOPPED},
	}
	instanceStateDefs[InstanceState_STOPPED] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_STARTING, InstanceState_SHUTTINGDOWN, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_RUNNING, InstanceState_TERMINATED},
	}
	instanceStateDefs[InstanceState_REBOOTING] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_RUNNING, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_RUNNING},
	}
	instanceStateDefs[InstanceState_SHUTTINGDOWN] = &stateDef{
		Nexts: []InstanceState_State{InstanceState_TERMINATED, InstanceState_FAILED},
		Goals: []InstanceState_State{InstanceState_TERMINATED},
	}
	instanceStateDefs[InstanceState_TERMINATED] = &stateDef{
		Nexts: []InstanceState_State{},
		Goals: []InstanceState_State{},
	}
	instanceStateDefs[InstanceState_FAILED] = &stateDef{
		Nexts: []InstanceState_State{},
		Goals: []InstanceState_State{},
	}
}

type instances struct {
	base
}

func Instances(ctx context.Context) InstanceOps {
	return &instances{base{ctx: ctx}}
}

func (i *instances) Create(n *Instance) (*Instance, error) {
	initState := &InstanceState{
		State: InstanceState_REGISTERED,
	}
	n.LastState = initState

	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID(fmt.Sprintf("/%s/i-", instancesBaseKey), n)
	if err != nil {
		return nil, err
	}
	if err := bk.Find(nkey, n); err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	if err = bk.Backend().Create(fmt.Sprintf("%s/state", nkey), []byte{}); err != nil {
		return nil, err
	}
	_, err = bk.CreateWithID(fmt.Sprintf("%s/state/state-", nkey), initState)
	if err != nil {
		return nil, err
	}

	if err = bk.Backend().Create(fmt.Sprintf("%s/failure-messages", nkey), []byte{}); err != nil {
		return nil, err
	}

	return n, nil
}

func (i *instances) Update(instance *Instance) error {
	if instance.Id == "" {
		return ErrInvalidID
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	err = bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, instance.Id), instance)
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
	n := &Instance{}
	if err := bk.Find(fmt.Sprintf("/%s/%s", instancesBaseKey, id), n); err != nil {
		return nil, err
	}
	n.Id = id
	return n, nil
}

func (i *instances) AddFailureMessage(id string, failureMessage FailureMessage_ErrorType) error {
	instance, err := i.FindByID(id)
	if err != nil {
		return err
	}
	bk, err := i.connection()
	if err != nil {
		return err
	}
	failedAt, _ := ptypes.TimestampProto(time.Now())
	nFailure := &FailureMessage{
		ErrorType: failureMessage,
		FailedAt:  failedAt,
	}
	instance.LatestFailure = nFailure
	_, err = bk.CreateWithID(fmt.Sprintf("%s/%s/failure-messages/failure-", instancesBaseKey, id), nFailure)
	if err != nil {
		return err
	}
	return bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, id), instance)
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
	n := &InstanceState{}
	if err := bk.Find(key, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (i *instances) UpdateState(id string, next InstanceState_State) error {
	instance, err := i.FindByID(id)
	if err != nil {
		return err
	}
	if err := instance.LastState.ValidateNextState(next); err != nil {
		return err
	}
	nstate := &InstanceState{
		State: next,
	}
	instance.LastState = nstate

	bk, err := i.connection()
	if err != nil {
		return err
	}
	_, err = bk.CreateWithID(fmt.Sprintf("/%s/%s/state/state-", instancesBaseKey, id), nstate)
	if err != nil {
		return err
	}
	return bk.Update(fmt.Sprintf("/%s/%s", instancesBaseKey, id), instance)
}

func (i *instances) FilterByState(state InstanceState_State) ([]*Instance, error) {
	res := []*Instance{}
	err := i.Filter(0, func(inst *Instance) int {
		if inst.GetLastState().State == state {
			res = append(res, inst)
		}
		return len(res)
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (i *instances) Filter(limit int, cb func(*Instance) int) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}
	keys, err := bk.Keys(fmt.Sprintf("/%s", instancesBaseKey))
	if err != nil {
		return err
	}
	for keys.Next() {
		instance, err := i.FindByID(keys.Value())
		if err != nil {
			return err
		}
		if limit > 0 && cb(instance) > limit {
			break
		} else {
			cb(instance)
		}
	}
	return nil
}

func (i *instances) fmtKey(id string, subkeys ...string) string {
	key := path.Join("/", instancesBaseKey, id)
	if len(subkeys) > 0 {
		key += "/" + path.Join(subkeys...)
	}
	return path.Clean(key)
}

func (i *instances) WaitStateUpdate(id string) (*InstanceState, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	watcher, ok := bk.(backend.ModelWatcher)
	if !ok {
		return nil, errors.Errorf("%T does not support backend.ModelWatcher", bk)
	}
	ev, err := watcher.Watch(i.fmtKey(id, "/state"))
	if err != nil {
		return nil, err
	}
	if ev != backend.EventModified {
		return nil, errors.Errorf("Unexpected event: %s", ev)
	}

	lkey, err := bk.FindLastKey(i.fmtKey(id, "/state/state-"))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed FindLaskKey(%s)", i.fmtKey(id, "/state/state-"))
	}
	res := &InstanceState{}
	if err := bk.Find(lkey, res); err != nil {
		return nil, errors.Wrapf(err, "Failed Find(%s)", lkey)
	}
	return res, nil
}

func (i *Instance) ResourceTemplate() ResourceTemplate {
	return GetResourceTemplate(i.GetTemplate())
}

func (i *InstanceState) ValidateNextState(next InstanceState_State) error {
	if i.GetState() == next {
		return fmt.Errorf("Instance is already %s", i.GetState().String())
	}
	if InstanceTerminalStates.Contains(i.GetState()) {
		return fmt.Errorf("Instance is already in terminal state: %s", i.GetState())
	}
	if instanceStateDefs[i.GetState()].Nexts.Contains(next) {
		return nil
	}

	return fmt.Errorf("Invalid next state: %s -> %s",
		i.GetState().String(),
		next.String(),
	)
}

type InstanceStateSlice []InstanceState_State

func (s InstanceStateSlice) Contains(state InstanceState_State) bool {
	for _, st := range s {
		if st == state {
			return true
		}
	}
	return false
}

var InstanceTerminalStates InstanceStateSlice = []InstanceState_State{
	InstanceState_TERMINATED,
	InstanceState_FAILED,
}

func (i *InstanceState) ValidateGoalState(goal InstanceState_State) error {
	if i.GetState() == goal {
		return fmt.Errorf("Instance is already in goal state: %s", goal)
	}
	if InstanceTerminalStates.Contains(i.GetState()) {
		return fmt.Errorf("Instance is already in terminal state: %s", i.GetState())
	}
	if instanceStateDefs[i.GetState()].Goals.Contains(goal) {
		return nil
	}

	return fmt.Errorf("Invalid goal state: %s -> %s",
		i.GetState().String(),
		goal.String(),
	)
}

var instanceConsoleStates InstanceStateSlice = []InstanceState_State{
	InstanceState_RUNNING,
}

func (i *InstanceState) ReadyForConsole() error {
	if instanceConsoleStates.Contains(i.GetState()) {
		return nil
	}
	return errors.New("Instance is not active to return console")
}
