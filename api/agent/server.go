package agent

import (
	"net"

	"golang.org/x/net/context"
	"github.com/axsh/openvdc/model"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/ptypes/empty"
)

//go:generate protoc -I../../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../../proto/agent.proto

type AgentAPIServer struct {
	listener  net.Listener
	server    *grpc.Server
}

func NewAgentAPIServer(r *model.ComputingResources) *AgentAPIServer {
	s := &AgentAPIServer{
		server: grpc.NewServer(),
	}
	RegisterResourceCollectorServer(s.server, &AgentAPI{api: s, resources: r})
	return s
}

func (s *AgentAPIServer) Serve(listen net.Listener) error {
	s.listener = listen
	return s.server.Serve(listen)
}

func (s *AgentAPIServer) Stop() {
	s.server.Stop()
	s.listener = nil
}

func (s *AgentAPIServer) GracefulStop() {
	s.server.GracefulStop()
	s.listener = nil
}


func (s *AgentAPIServer) Listener() net.Listener {
	return s.listener
}

type AgentAPI struct {
	api       *AgentAPIServer
	resources *model.ComputingResources
}


func (api *AgentAPI) GetResources(ctx context.Context, e *empty.Empty) (*model.ComputingResources, error) {
	return api.resources, nil
}
