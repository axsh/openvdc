package executor

import (
	"net"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//go:generate protoc -I../../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../../proto/executor.proto

type ExecutorAPIServer struct {
	server         *grpc.Server
	listener       net.Listener
	modelStoreAddr backend.ConnectionAddress
}

func NewExecutorAPIServer(modelAddr backend.ConnectionAddress, ctx context.Context) *ExecutorAPIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr, ctx)),
		grpc.StreamInterceptor(model.GrpcStreamInterceptor(modelAddr, ctx)),
	}
	s := &ExecutorAPIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
	}

	return s
}

func (s *ExecutorAPIServer) Serve(listen net.Listener) error {
	s.listener = listen
	return s.server.Serve(listen)
}

func (s *ExecutorAPIServer) Stop() {
	s.server.Stop()
	s.listener = nil
}

func (s *ExecutorAPIServer) GracefulStop() {
	s.server.GracefulStop()
	s.listener = nil
}

func (s *ExecutorAPIServer) Listener() net.Listener {
	return s.listener
}
