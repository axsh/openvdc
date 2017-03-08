package api

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//go:generate protoc -I../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../proto/v1.proto

type APIServer struct {
	server           *grpc.Server
	modelStoreAddr   backend.ConnectionAddress
	scheduler        sched.SchedulerDriver
	mesosMasterAddr  *mesos.Address
	mMesosMasterAddr sync.Mutex
}

func NewAPIServer(modelAddr backend.ConnectionAddress, driver sched.SchedulerDriver, ctx context.Context) *APIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr, ctx)),
		grpc.StreamInterceptor(model.GrpcStreamInterceptor(modelAddr, ctx)),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
		scheduler:      driver,
	}

	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterInstanceConsoleServer(s.server, &InstanceConsoleAPI{api: s})
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

func (s *APIServer) GetMesosMasterAddr() *mesos.Address {
	s.mMesosMasterAddr.Lock()
	defer s.mMesosMasterAddr.Unlock()
	return s.mesosMasterAddr
}

// detector.MasterChanged interface
func (s *APIServer) OnMasterChanged(info *mesos.MasterInfo) {
	if info == nil {
		log.Error("Lost mesos master")
	} else {
		s.mMesosMasterAddr.Lock()
		defer s.mMesosMasterAddr.Unlock()
		if info.GetAddress() == nil {
			// IP:Port address is specified to --master
			// --master=192.168.56.150:5050 or --master=localhost:5050
			s.mesosMasterAddr = &mesos.Address{
				Ip:       info.Hostname,
				Port:     proto.Int32(int32(info.GetPort())),
				Hostname: info.Hostname,
			}
		} else {
			// zk:// URL is specified to --master
			// --master=zk://192.168.56.150/mesos or --master=zk://localhost/mesos
			s.mesosMasterAddr = info.GetAddress()
		}
	}
}
