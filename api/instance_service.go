package api

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	util "github.com/mesos/mesos-go/mesosutil"
	"golang.org/x/net/context"
)

type InstanceAPI struct {
	api *APIServer
}

func (s *InstanceAPI) Create(ctx context.Context, in *CreateRequest) (*CreateReply, error) {
	if in.GetResourceId() == "" {
		return nil, fmt.Errorf("Invalid Resource ID")
	}
	r, err := model.Resources(ctx).FindByID(in.GetResourceId())
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	if r.GetState() == model.Resource_UNREGISTERED {
		log.WithFields(log.Fields{
			"resource_id": in.GetResourceId(),
			"state":       r.GetState().String(),
		}).Error("Cannot use unregistered resource")

		return nil, fmt.Errorf("Cannot use unregistered resource")
	}

	inst, err := model.Instances(ctx).Create(&model.Instance{
		ResourceId: r.GetId(),
	})
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &CreateReply{InstanceId: inst.Id}, nil
}

func (s *InstanceAPI) Start(ctx context.Context, in *StartRequest) (*StartReply, error) {
	if in.GetInstanceId() == "" {
		return nil, fmt.Errorf("Invalid Instance ID")
	}
	inst, err := model.Instances(ctx).FindByID(in.GetInstanceId())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}
	lastState := inst.GetLastState()
	switch lastState.GetState() {
	case model.InstanceState_REGISTERED:
		if err := model.Instances(ctx).UpdateState(in.GetInstanceId(), model.InstanceState_QUEUED); err != nil {
			log.WithError(err).Error()
			return nil, err
		}
	case model.InstanceState_STOPPED:
		if err := s.sendCommand(ctx, "start", in.GetInstanceId()); err != nil {
			log.WithError(err).Error("Failed to sendCommand(start)")
			return nil, err
		}

	case model.InstanceState_RUNNING:
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error("Instance is already running")

		return nil, fmt.Errorf("Instance is already running")

	case model.InstanceState_TERMINATED:
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error("Cannot start terminated instance")

		return nil, fmt.Errorf("Cannot start terminated instance")
	default:
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error("Unexpected instance state")
		// TODO: Investigate gRPC error response
		return nil, fmt.Errorf("Unexpected instance state")
	}
	// TODO: Tell the scheduler there is a queued item to get next offer eagerly.
	return &StartReply{InstanceId: in.GetInstanceId()}, nil
}

func (s *InstanceAPI) Run(ctx context.Context, in *ResourceRequest) (*RunReply, error) {
	resourceAPI := &ResourceAPI{api: s.api}
	res0, err := resourceAPI.Register(ctx, in)
	if err != nil {
		log.WithError(err).Error("Failed InstanceAPI.Run at ResourceAPI.Register")
		return nil, err
	}
	resourceID := res0.GetID()
	res1, err := s.Create(ctx, &CreateRequest{ResourceId: resourceID})
	if err != nil {
		log.WithError(err).Error("Failed InstanceAPI.Run at Create")
		return nil, err
	}
	res2, err := s.Start(ctx, &StartRequest{InstanceId: res1.GetInstanceId()})
	if err != nil {
		log.WithError(err).Error("Failed InstanceAPI.Run at Start")
		return nil, err
	}
	return &RunReply{InstanceId: res2.GetInstanceId(), ResourceId: resourceID}, nil
}

func (s *InstanceAPI) Stop(ctx context.Context, in *StopRequest) (*StopReply, error) {

	if in.GetInstanceId() == "" {
		return nil, fmt.Errorf("Invalid Instance ID")
	}

	inst, err := model.Instances(ctx).FindByID(in.GetInstanceId())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}

	if inst.GetLastState().GetState() != model.InstanceState_RUNNING {
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       inst.GetLastState().GetState(),
		}).Error("Instance is not running")

		return nil, fmt.Errorf("Instance is not running")
	}

	instanceID := in.InstanceId
	if err := s.sendCommand(ctx, "stop", instanceID); err != nil {
		log.WithError(err).Error("Failed sendCommand(stop)")
		return nil, err
	}

	return &StopReply{InstanceId: instanceID}, nil
}

func (s *InstanceAPI) Destroy(ctx context.Context, in *DestroyRequest) (*DestroyReply, error) {

	instanceID := in.InstanceId

	if instanceID == "" {
		return nil, fmt.Errorf("Invalid Instance ID")
	}

	inst, err := model.Instances(ctx).FindByID(in.GetInstanceId())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}

	lastState := inst.GetLastState()
	switch lastState.GetState() {
	case model.InstanceState_TERMINATED:
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error("Instance is already terminated")

		return nil, fmt.Errorf("Instance is already terminated")
	default:

		if err := s.sendCommand(ctx, "destroy", instanceID); err != nil {
			log.WithError(err).Error("Failed sendCommand(destroy)")
			return nil, err
		}
	}

	return &DestroyReply{InstanceId: instanceID}, nil
}

func (s *InstanceAPI) Console(ctx context.Context, in *ConsoleRequest) (*ConsoleReply, error) {

	instanceID := in.InstanceId
	if err := s.sendCommand(ctx, "console", instanceID); err != nil {
		log.WithError(err).Error("Failed sendCommand(console)")
		return nil, err
	}

	return &ConsoleReply{InstanceId: instanceID}, nil
}

func (s *InstanceAPI) sendCommand(ctx context.Context, cmd string, instanceID string) error {
	inst, err := model.Instances(ctx).FindByID(instanceID)
	if err != nil {
		return err
	}
	// Fetch associated resource to the instance
	res, err := inst.Resource(ctx)
	if err != nil {
		return err
	}
	//There might be a better way to do this, but for now the AgentID is set through an environment variable.
	//Example: export AGENT_ID="81fd8c72-3261-4ce9-95c8-7fade4b290ad-S0"
	slaveID, ok := os.LookupEnv("AGENT_ID")
	if !ok {
		slaveID = inst.SlaveId
	}

	_, err = s.api.scheduler.SendFrameworkMessage(
		util.NewExecutorID(fmt.Sprintf("vdc-hypervisor-%s", res.ResourceTemplate().ResourceName())),
		util.NewSlaveID(slaveID),
		fmt.Sprintf("%s_%s", cmd, instanceID),
	)
	return err
}

func (s *InstanceAPI) Show(ctx context.Context, in *InstanceIDRequest) (*InstanceReply, error) {
	// in.Key takes nil possibly.
	if in.GetKey() == nil {
		log.Error("Invalid instance identifier")
		return nil, fmt.Errorf("Invalid instance identifier")
	}

	// TODO: handle the case for in.GetName() is received.
	instance, err := model.Instances(ctx).FindByID(in.GetID())
	if err != nil {
		log.WithError(err).WithField("key", in.GetID()).Error("Failed Instances.FindByID")
		return nil, err
	}
	return &InstanceReply{ID: instance.GetId(), Instance: instance}, nil
}

type InstanceConsoleAPI struct {
	api *APIServer
}

func (i *InstanceConsoleAPI) Attach(stream InstanceConsole_AttachServer) error {
	in, err := stream.Recv()
	if err != nil {
		return err
	}
	instanceID := in.GetInstanceId()
	if instanceID == "" {
		// Return error if no instance ID is set to the first request.
		return fmt.Errorf("instance_id not found")
	}
	_, err = model.Instances(stream.Context()).FindByID(instanceID)
	if err != nil {
		return err
	}
	return nil
}
