package api

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type APIOffer chan *RunRequest

type APIServer struct {
	server    *grpc.Server
	offerChan APIOffer
}

func NewAPIServer(c APIOffer) *APIServer {
	s := &APIServer{
		server:    grpc.NewServer(),
		offerChan: c,
	}
	RegisterInstanceServer(s.server, &RemoteAPI{api: s})
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

type RemoteAPI struct {
	api *APIServer
}

func (s *RemoteAPI) Run(ctx context.Context, in *RunRequest) (*RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	s.api.offerChan <- in
	return &RunReply{}, nil
}
