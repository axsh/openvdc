package main

import (
	"log"
	"os"

	pb "github.com/axsh/openvdc/proto"
	fp "gopkg.in/alecthomas/kingpin.v2"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	addrFlag = fp.Flag("addr", "server address host:post").Default("localhost:5000").String()
        task = fp.Flag("task", "Call task: lxc-create, lxc-destroy").String()
)

func main() {

	fp.Parse()

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
