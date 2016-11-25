package api

import (
	"errors"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/model"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	sched "github.com/mesos/mesos-go/scheduler"

	util "github.com/mesos/mesos-go/mesosutil"
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

func (s *InstanceAPI) StopTask(ctx context.Context, in *StopTaskRequest) (*StopTaskReply, error) {

	hostName := in.HostName

	//TODO: Don't hardcode the ID's.
	theDriver.SendFrameworkMessage(
		util.NewExecutorID("vdc-hypervisor-null"),
		util.NewSlaveID("be590de8-83c0-47f5-9e4a-14f5326c240b-S0"),
		"destroy_"+hostName,
	)

	return &StopTaskReply{InstanceId: "test"}, nil
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
	if err := model.Instances(ctx).UpdateState(in.GetInstanceId(), model.Instance_QUEUED); err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	// TODO: Tell the scheduler there is a queued item to get next offer eagerly.
	return &StartReply{InstanceId: in.GetInstanceId()}, nil
}

func (s *InstanceAPI) Run(ctx context.Context, in *RunRequest) (*RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	inst, err := model.Instances(ctx).Create(&model.Instance{})
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	in.HostName = inst.Id
	s.api.offerChan <- in
	return &RunReply{InstanceId: inst.Id}, nil
}

type ResourceAPI struct {
	api *APIServer
}

var ErrTemplateUndefined = errors.New("Template is undefined")

func (s *ResourceAPI) Register(ctx context.Context, in *ResourceRequest) (*ResourceReply, error) {
	r := &model.Resource{}
	switch x := in.Template.(type) {
	case *ResourceRequest_None:
		r.Type = model.ResourceType_NONE
		r.Template = &model.Resource_None{None: x.None}
	case *ResourceRequest_Vm:
		r.Type = model.ResourceType_VM
		r.Template = &model.Resource_Vm{Vm: x.Vm}
	case nil:
		log.WithError(ErrTemplateUndefined).Error("template parameter is nil")
		return nil, ErrTemplateUndefined
	}
	resource, err := model.Resources(ctx).Create(r)
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &ResourceReply{ID: resource.GetId()}, nil
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
