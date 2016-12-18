package api

import (
	"net"

	"github.com/axsh/openvdc/model"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
	scheduler      sched.SchedulerDriver
}

func NewAPIServer(modelAddr string, driver sched.SchedulerDriver, ctx context.Context) *APIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr, ctx)),
		grpc.StreamInterceptor(model.GrpcStreamInterceptor(modelAddr, ctx)),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
		scheduler:      driver,
	}

	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
	RegisterInstanceConsoleServer(s.server, &InstanceConsoleAPI{api: s})
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
