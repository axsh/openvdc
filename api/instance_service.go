package api

import (
	"fmt"
	"os"
	"strings"

	mlog "github.com/ContainX/go-mesoslog/mesoslog"
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	util "github.com/mesos/mesos-go/mesosutil"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type InstanceAPI struct {
	api *APIServer
}

func (s *InstanceAPI) Create(ctx context.Context, in *CreateRequest) (*CreateReply, error) {
	inst, err := model.Instances(ctx).Create(&model.Instance{
		Template: in.GetTemplate(),
	})
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &CreateReply{InstanceId: inst.Id}, nil
}

func checkSupportAPI(t *model.Template, ctx context.Context) error {
	rt, ok := t.Item.(model.ResourceTemplate)
	if !ok {
		return errors.Errorf("Invalid type: %T", t.Item)
	}
	h, ok := handlers.FindByType(rt.ResourceName())
	if !ok {
		return errors.Errorf("Unknown resource name: %s", rt.ResourceName())
	}
	md, _ := metadata.FromContext(ctx)
	if !h.IsSupportAPI(md["fullmethod"][0]) {
		return errors.Errorf("%s is not supported: %T", md["fullmethod"][0], t.Item)
	}
	return nil
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
	if err := checkSupportAPI(inst.GetTemplate(), ctx); err != nil {
		return nil, err
	}
	lastState := inst.GetLastState()
	flog := log.WithFields(log.Fields{
		"instance_id": in.GetInstanceId(),
		"state":       lastState.String(),
	})
	switch lastState.GetState() {
	case model.InstanceState_REGISTERED:
		if err := lastState.ValidateGoalState(model.InstanceState_QUEUED); err != nil {
			flog.Error(err)
			// TODO: Investigate gRPC error response
			return nil, err
		}
		if err := model.Instances(ctx).UpdateState(in.GetInstanceId(), model.InstanceState_QUEUED); err != nil {
			flog.Error(err)
			return nil, err
		}
	case model.InstanceState_STOPPED:
		if err := lastState.ValidateGoalState(model.InstanceState_RUNNING); err != nil {
			flog.Error(err)
			// TODO: Investigate gRPC error response
			return nil, err
		}
		if err := s.sendCommand(ctx, "start", in.GetInstanceId()); err != nil {
			flog.WithError(err).Error("Failed to sendCommand(start)")
			return nil, err
		}
	default:
		flog.Fatal("BUGON: Detected un-handled state")
	}
	// TODO: Tell the scheduler there is a queued item to get next offer eagerly.
	return &StartReply{InstanceId: in.GetInstanceId()}, nil
}

func (s *InstanceAPI) Run(ctx context.Context, in *CreateRequest) (*RunReply, error) {
	if err := checkSupportAPI(in.GetTemplate(), ctx); err != nil {
		return nil, err
	}
	res1, err := s.Create(ctx, &CreateRequest{Template: in.GetTemplate()})
	if err != nil {
		log.WithError(err).Error("Failed InstanceAPI.Run at Create")
		return nil, err
	}
	res2, err := s.Start(ctx, &StartRequest{InstanceId: res1.GetInstanceId()})
	if err != nil {
		log.WithError(err).Error("Failed InstanceAPI.Run at Start")
		return nil, err
	}
	return &RunReply{InstanceId: res2.GetInstanceId()}, nil
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
	if err := checkSupportAPI(inst.GetTemplate(), ctx); err != nil {
		return nil, err
	}

	if err := inst.GetLastState().ValidateGoalState(model.InstanceState_STOPPED); err != nil {
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       inst.GetLastState().GetState(),
		}).Error(err)

		return nil, err
	}

	instanceID := in.InstanceId
	if err := s.sendCommand(ctx, "stop", instanceID); err != nil {
		log.WithError(err).Error("Failed sendCommand(stop)")
		return nil, err
	}

	return &StopReply{InstanceId: instanceID}, nil
}

func (s *InstanceAPI) Reboot(ctx context.Context, in *RebootRequest) (*RebootReply, error) {

	if in.GetInstanceId() == "" {
		return nil, fmt.Errorf("Invalid Instance ID")
	}

	inst, err := model.Instances(ctx).FindByID(in.GetInstanceId())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}
	if err := checkSupportAPI(inst.GetTemplate(), ctx); err != nil {
		return nil, err
	}
	if err := inst.GetLastState().ValidateGoalState(model.InstanceState_STOPPED); err != nil {
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       inst.GetLastState().GetState(),
		}).Error(err)

		return nil, err
	}

	instanceID := in.InstanceId
	if err := s.sendCommand(ctx, "reboot", instanceID); err != nil {
		log.WithError(err).Error("Failed sendCommand(reboot)")
		return nil, err
	}

	return &RebootReply{InstanceId: instanceID}, nil
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
	if err := lastState.ValidateGoalState(model.InstanceState_TERMINATED); err != nil {
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error(err)
		return nil, err
	}

	currentState := inst.GetLastState().GetState()

	if currentState == model.InstanceState_REGISTERED {
		err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_TERMINATED)
		if err != nil {
			log.WithError(err).Error("Failed to update instance state.")
		}
	} else {
		if err := s.sendCommand(ctx, "destroy", instanceID); err != nil {
			log.WithError(err).Error("Failed sendCommand(destroy)")
			return nil, err
		}
	}

	return &DestroyReply{InstanceId: instanceID}, nil
}

func (s *InstanceAPI) Console(ctx context.Context, in *ConsoleRequest) (*ConsoleReply, error) {

	instanceID := in.InstanceId
	if instanceID == "" {
		return nil, fmt.Errorf("Invalid Instance ID")
	}

	inst, err := model.Instances(ctx).FindByID(in.GetInstanceId())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}
	if err := checkSupportAPI(inst.GetTemplate(), ctx); err != nil {
		return nil, err
	}
	lastState := inst.GetLastState()
	if err := lastState.ReadyForConsole(); err != nil {
		log.WithFields(log.Fields{
			"instance_id": in.GetInstanceId(),
			"state":       lastState.String(),
		}).Error(err)
		return nil, err
	}
	node := &model.ExecutorNode{}
	if err := model.Cluster(ctx).Find(inst.GetSlaveId(), node); err != nil {
		log.WithError(err).WithField("instance_id", in.GetInstanceId()).Error("Failed to find the instance")
		return nil, err
	}

	return &ConsoleReply{
		InstanceId: instanceID,
		Type:       node.Console.Type,
		Address:    node.Console.BindAddr,
	}, nil
}

func (s *InstanceAPI) sendCommand(ctx context.Context, cmd string, instanceID string) error {
	inst, err := model.Instances(ctx).FindByID(instanceID)
	if err != nil {
		return err
	}
	//There might be a better way to do this, but for now the AgentID is set through an environment variable.
	//Example: export AGENT_ID="81fd8c72-3261-4ce9-95c8-7fade4b290ad-S0"
	slaveID, ok := os.LookupEnv("AGENT_ID")
	if !ok {
		slaveID = inst.SlaveId
	}

	hypervisorName := strings.TrimPrefix(inst.ResourceTemplate().ResourceName(), "vm/")
	_, err = s.api.scheduler.SendFrameworkMessage(
		util.NewExecutorID(fmt.Sprintf("vdc-hypervisor-%s", hypervisorName)),
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

func (s *InstanceAPI) List(ctx context.Context, in *InstanceListRequest) (*InstanceListReply, error) {
	page := &InstanceListRequest_PageRequest{
		Limit:  1000,
		Offset: 0,
	}
	if in.Page != nil {
		page = in.Page
	}

	results := []*InstanceListReply_InstanceListItem{}
	err := model.Instances(ctx).Filter(int(page.Limit), func(i *model.Instance) int {
		found := false
		if in.Filter == nil {
			found = true
		} else {
			found = in.Filter.State == i.GetLastState().State
		}

		if found {
			results = append(results, &InstanceListReply_InstanceListItem{
				Id:    i.Id,
				State: i.GetLastState().State,
			})
		}
		return len(results)
	})
	if err != nil {
		log.WithError(err).Error("Failed Instances.Filter")
		return nil, err
	}
	return &InstanceListReply{
		Page: &InstanceListReply_PageReply{
			Total: int32(len(results)),
			Limit: page.Limit,
		},
		Items: results,
	}, nil
}

func (s *InstanceAPI) Log(in *InstanceLogRequest, stream Instance_LogServer) error {
	inst, err := model.Instances(stream.Context()).FindByID(in.Target.GetID())
	if err != nil {
		log.WithError(err).WithField("instance_id", in.Target.GetID()).Error("Failed to find the instance")
		return err
	}
	if err := checkSupportAPI(inst.GetTemplate(), stream.Context()); err != nil {
		return err
	}
	masterAddr := s.api.GetMesosMasterAddr()
	if masterAddr == nil {
		return errors.New("Mesos master address is not detected")
	}
	cl, err := mlog.NewMesosClientWithOptions(
		masterAddr.GetIp(),
		int(masterAddr.GetPort()),
		&mlog.MesosClientOptions{SearchCompletedTasks: false, ShowLatestOnly: true})
	if err != nil {
		log.WithError(err).Error("Couldn't connect to Mesos master: ", masterAddr)
		return errors.Wrap(err, "mlog.NewMesosClientWithOptions")
	}

	taskID := fmt.Sprintf("VDC_%s", in.Target.GetID())
	result, err := cl.GetLog(taskID, mlog.STDERR, "")
	if err != nil {
		log.WithError(err).Error("Error fetching log")
		return errors.Wrap(err, "cl.GetLog")
	}

	for _, log := range result {
		err := stream.Send(&InstanceLogReply{
			Line: []string{log.Log},
		})
		if err != nil {
			return errors.Wrap(err, "stream.Send")
		}
	}
	return nil
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
