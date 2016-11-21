package api

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/model"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	sched "github.com/mesos/mesos-go/scheduler"

	util "github.com/mesos/mesos-go/mesosutil"
)

type APIOffer chan *RunRequest
var theDriver sched.SchedulerDriver

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
	offerChan      APIOffer
}

func NewAPIServer(c APIOffer, modelAddr string, driver sched.SchedulerDriver) *APIServer {
	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr)),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		offerChan:      c,
		modelStoreAddr: modelAddr,
	}

	theDriver = driver

	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
	return s
}

func (s *InstanceAPI) StopTask(ctx context.Context, in *StopTaskRequest) (*StopTaskReply, error) {

        //TODO: Don't hardcode the ID's.
        theDriver.SendFrameworkMessage(
                util.NewExecutorID("vdc-hypervisor-null"),
                util.NewSlaveID("3870d2ed-0ef0-40b4-a6f3-bea46d43d5e1-S0"),
                "destroy",
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

func (s *InstanceAPI) Run(ctx context.Context, in *RunRequest) (*RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	inst, err := model.Instances(ctx).Create(&model.Instance{})
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	s.api.offerChan <- in
	return &RunReply{InstanceId: inst.Id}, nil
}

type ResourceAPI struct {
	api *APIServer
}

func (s *ResourceAPI) Register(ctx context.Context, in *ResourceRequest) (*ResourceReply, error) {
	resource, err := model.Resources(ctx).Create(&model.Resource{})
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
