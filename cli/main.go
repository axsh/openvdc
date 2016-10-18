package main

import (
	"flag"
	"log"
	"os"

	pb "github.com/axsh/openvdc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	addrFlag = flag.String("addr", "localhost:5000", "server address host:post")
)

func main() {
	conn, err := grpc.Dial(*addrFlag, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Connection error: %v", err)
		os.Exit(1)
	}

	defer conn.Close()

	c := pb.NewInstanceClient(conn)

	resp, err := c.Run(context.Background(), &pb.RunRequest{})
	if err != nil {
		log.Fatalf("RPC error: %v", err)
	}
	log.Println(resp)
}
