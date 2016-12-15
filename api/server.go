package api

import (
	"net"

	"github.com/axsh/openvdc/model"
	"google.golang.org/grpc"

	sched "github.com/mesos/mesos-go/scheduler"
)

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
	scheduler      sched.SchedulerDriver
}

func NewAPIServer(modelAddr string, driver sched.SchedulerDriver) *APIServer {
	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr)),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
		scheduler:      driver,
	}

	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
	return s
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
