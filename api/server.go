package api

import (
	"net"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"google.golang.org/grpc"

	sched "github.com/mesos/mesos-go/scheduler"
)

//go:generate protoc -I../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../proto/v1.proto

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr backend.ConnectionAddress
	scheduler      sched.SchedulerDriver
}

func NewAPIServer(modelAddr backend.ConnectionAddress, driver sched.SchedulerDriver) *APIServer {
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
