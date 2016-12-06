package api

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/axsh/openvdc/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	sched "github.com/mesos/mesos-go/scheduler"
	util "github.com/mesos/mesos-go/mesosutil"
	log "github.com/Sirupsen/logrus"
)

var theDriver sched.SchedulerDriver

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
}

func NewAPIServer(modelAddr string, driver sched.SchedulerDriver) *APIServer {
	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr)),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
	}

	theDriver = driver

	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
	return s
}

func (s *InstanceAPI) Stop(ctx context.Context, in *StopRequest) (*StopReply, error) {

	instanceID := in.InstanceId
	sendCommand("stop", instanceID)

	return &StopReply{InstanceId: instanceID + " stopped."}, nil
}

func (s *InstanceAPI) Destroy(ctx context.Context, in *DestroyRequest) (*DestroyReply, error) {

        instanceID := in.InstanceId
        sendCommand("destroy", instanceID)

        return &DestroyReply{InstanceId: instanceID + " destroyed."}, nil
}

func (s *InstanceAPI) Console(ctx context.Context, in *ConsoleRequest) (*ConsoleReply, error) {

        instanceID := in.InstanceId
        sendCommand("console", instanceID)

        return &ConsoleReply{InstanceId: instanceID}, nil
}

func sendCommand(cmd string, id string) {
	
	if os.Getenv("AGENT_ID") == "" {
                log.Errorln("AGENT_ID env variable needs to be set. Example: AGENT_ID=81fd8c72-3261-4ce9-95c8-7fade4b290ad-S0")
        } else {
                //There might be a better way to do this, but for now the AgentID is set through an environment variable.
                //Example: export AGENT_ID="81fd8c72-3261-4ce9-95c8-7fade4b290ad-S0"
                theDriver.SendFrameworkMessage(
                        util.NewExecutorID("vdc-hypervisor-null"),
                        util.NewSlaveID(os.Getenv("AGENT_ID")),
                        cmd + "_" + id,
                )
        }
}

func (s *APIServer) Serve(listen net.Listener) error {
	return s.server.Serve(listen)
}

func (s *APIServer) Stop() {
	s.server.Stop()
}

func (s *APIServer) GracefulStop() {
	s.server.GracefulStop()
}

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
	if err := model.Instances(ctx).UpdateState(in.GetInstanceId(), model.InstanceState_QUEUED); err != nil {
		log.WithError(err).Error()
		return nil, err
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

type ResourceAPI struct {
	api *APIServer
}

var ErrTemplateUndefined = errors.New("Template is undefined")
var ErrUnknownTemplate = errors.New("Unknown template type")

func (s *ResourceAPI) Register(ctx context.Context, in *ResourceRequest) (*ResourceReply, error) {
	r := &model.Resource{
		TemplateUri: in.GetTemplateUri(),
	}
	switch x := in.Template.(type) {
	case *ResourceRequest_None:
		r.Type = model.ResourceType_RESOURCE_NONE
		r.Template = &model.Resource_None{None: x.None}
	case *ResourceRequest_Lxc:
		r.Type = model.ResourceType_RESOURCE_LXC
		r.Template = &model.Resource_Lxc{Lxc: x.Lxc}
	case *ResourceRequest_Null:
		r.Type = model.ResourceType_RESOURCE_NULL
		r.Template = &model.Resource_Null{Null: x.Null}
	case nil:
		log.WithError(ErrTemplateUndefined).Error("template parameter is nil")
		return nil, ErrTemplateUndefined
	default:
		log.Error("Unsupported template type")
		return nil, ErrUnknownTemplate
	}
	resource, err := model.Resources(ctx).Create(r)
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &ResourceReply{ID: resource.GetId(), Resource: resource}, nil
}
func (s *ResourceAPI) Unregister(ctx context.Context, in *ResourceIDRequest) (*ResourceReply, error) {
	// in.Key takes nil possibly.
	if in.GetKey() == nil {
		log.Error("Invalid resource identifier")
		return nil, fmt.Errorf("Invalid resource identifier")
	}
	// TODO: handle the case for in.GetName() is received.
	err := model.Resources(ctx).Destroy(in.GetID())
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	return &ResourceReply{ID: in.GetID()}, nil
}

func (s *ResourceAPI) Show(ctx context.Context, in *ResourceIDRequest) (*ResourceReply, error) {
	// in.Key takes nil possibly.
	if in.GetKey() == nil {
		log.Error("Invalid resource identifier")
		return nil, fmt.Errorf("Invalid resource identifier")
	}
	// TODO: handle the case for in.GetName() is received.
	resource, err := model.Resources(ctx).FindByID(in.GetID())
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &ResourceReply{ID: resource.GetId(), Resource: resource}, nil
}
