package main

import (
	"flag"
	"log"
	"net"
	"os"

	pb "github.com/axsh/openvdc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	addrFlag = flag.String("addr", ":5000", "Address host:post")
)

type server struct{}

func (s *server) Run(ctx context.Context, in *pb.RunRequest) (*pb.RunReply, error) {
	log.Printf("New Request: %v", in.String())
	return &pb.RunReply{}, nil
}

func main() {
	lis, err := net.Listen("tcp", *addrFlag)

	if err != nil {
		log.Fatalf("unknown bind address")
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterInstanceServer(s, &server{})
	s.Serve(lis)
}
