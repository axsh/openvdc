package main

import (
	"flag"
	"log"
	"net"

	pb "github.com/axsh/openvdc/proto"
	"github.com/mesos/mesos-go/detector"
	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	addrFlag       = flag.String("addr", ":5000", "Address host:port")
	zkMesosMasters = flag.String("zk", "zk://localhost:2181/mesos", "ZK Address URI")
)

type server struct{}

func (s *server) Run(ctx context.Context, in *pb.RunRequest) (*pb.RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	return &pb.RunReply{}, nil
}

type zkListener struct{}

func (l *zkListener) OnMasterChanged(info *mesos.MasterInfo) {
	if info == nil {
		log.Println("master lost")
	} else {
		log.Printf("master changed: Id %v Ip %v Hostname %v Port %v Version %v Pid %v\n",
			info.GetId(), info.GetIp(), info.GetHostname(), info.GetPort(), info.GetVersion(), info.GetPid())
	}
}

func (l *zkListener) UpdatedMasters(all []*mesos.MasterInfo) {
	for i, info := range all {
		log.Printf("master (%d): Id %v Ip %v Hostname %v Port %v Version %v Pid %v\n", i,
			info.GetId(), info.GetIp(), info.GetHostname(), info.GetPort(), info.GetVersion(), info.GetPid())
	}
}

func init() {
	flag.Parse()
}

func main() {
	lis, err := net.Listen("tcp", *addrFlag)
	if err != nil {
		log.Fatalln("unknown bind address: ", addrFlag)
	}

	md, err := detector.New(*zkMesosMasters)
	if err != nil {
		log.Fatalln("Failed to create ZK listener for Mesos masters: ", err)
	}
	err = md.Detect(&zkListener{})
	if err != nil {
		log.Fatalln("Faild to register ZK listener: ", err)
	}

	s := grpc.NewServer()
	pb.RegisterInstanceServer(s, &server{})
	s.Serve(lis)
}
