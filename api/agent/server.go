package agent

import (
	"net"

	"golang.org/x/net/context"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"google.golang.org/grpc"
	empty "github.com/golang/protobuf/ptypes/empty"
)

//go:generate protoc -I../../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../../proto/agent.proto

type ResourceCollectorAPIServer struct {
	listener  net.Listener
	server    *grpc.Server
}

func NewResourceCollectorAPIServer(ctx context.Context, modelAddr backend.ConnectionAddress, nodes map[string]*model.MonitorNode) *ResourceCollectorAPIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr, ctx)),
		grpc.StreamInterceptor(model.GrpcStreamInterceptor(modelAddr, ctx)),
	}

	s := &ResourceCollectorAPIServer{
		server: grpc.NewServer(sopts...),
	}
	RegisterResourceCollectorServer(s.server, &ResourceCollectorAPI{
		api:          s,
		monitorNodes: nodes,
	})
	return s
}

func (s *ResourceCollectorAPIServer) Serve(listen net.Listener) error {
	s.listener = listen
	return s.server.Serve(listen)
}

func (s *ResourceCollectorAPIServer) Stop() {
	s.server.Stop()
	s.listener = nil
}

func (s *ResourceCollectorAPIServer) GracefulStop() {
	s.server.GracefulStop()
	s.listener = nil
}

func (s *ResourceCollectorAPIServer) Listener() net.Listener {
	return s.listener
}

type ResourceCollectorAPI struct {
	api          *ResourceCollectorAPIServer
	monitorNodes map[string]*model.MonitorNode     
}

func (api *ResourceCollectorAPI) ReportResources(ctx context.Context, n *model.MonitorNode) (*empty.Empty, error) {
	api.monitorNodes[n.GetId()] = n
	return &empty.Empty{}, nil
}
