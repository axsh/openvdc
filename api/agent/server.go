package agent

import (
	"net"


	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//go:generate protoc -I../../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../../proto/agent.proto

type AgentAPIServer struct {
	listener net.Listener
	server   *grpc.Server
}

func NewAgentAPIServer() *AgentAPIServer {
	return nil
}

func (s *AgentAPIServer) Serve(listen net.Listener) error {
	s.listener = listen
	return s.server.Serve(listen)
}

func (s *AgentAPIServer) Stop() {
	s.Stop()
	s.listener = nil
}

func (s *AgentAPIServer) Listener() net.Listener {
	return s.listener
}
