package api

import (
	"log"
	"net"

	pb "github.com/axsh/openvdc/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type APIServer struct {
	server *grpc.Server
}

func NewAPIServer() *APIServer {
	s := grpc.NewServer()
	pb.RegisterInstanceServer(s, &RemoteAPI{})
	return &APIServer{
		server: s,
	}
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

type RemoteAPI struct{}

func (s *RemoteAPI) Run(ctx context.Context, in *pb.RunRequest) (*pb.RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	return &pb.RunReply{}, nil
}
