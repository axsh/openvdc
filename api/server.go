package api

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/scheduler"
	"github.com/gogo/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//go:generate protoc -I../proto -I${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ../proto/v1.proto

type APIServer struct {
	server            *grpc.Server
	modelStoreAddr    backend.ConnectionAddress
	scheduler         sched.SchedulerDriver
	mesosMasterAddr   *mesos.Address
	mMesosMasterAddr  sync.Mutex
	instanceScheduler scheduler.Schedule
}

type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStreamWithContext) Context() context.Context {
	return ss.ctx
}

func NewAPIServer(modelAddr backend.ConnectionAddress, driver sched.SchedulerDriver, instanceScheduler scheduler.Schedule, ctx context.Context) *APIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	// Nest UnaryInterceptor and StreamInterceptor
	modelInterceptor := model.GrpcInterceptor(modelAddr, ctx)
	insertFullMethod := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromContext(ctx)
		if ok {
			md = metadata.Join(md, metadata.Pairs("fullmethod", info.FullMethod))
		} else {
			log.Error("Failed metadata.FromContext")
		}
		return modelInterceptor(metadata.NewContext(ctx, md), req, info, handler)
	}
	modelStreamInterceptor := model.GrpcStreamInterceptor(modelAddr, ctx)
	insertFullMethodStream := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromContext(ss.Context())
		if ok {
			md = metadata.Join(md, metadata.Pairs("fullmethod", info.FullMethod))
		} else {
			log.Error("Failed metadata.FromContext")
		}
		ss = &serverStreamWithContext{ss, metadata.NewContext(ctx, md)}
		return modelStreamInterceptor(srv, ss, info, handler)
	}

	sopts := []grpc.ServerOption{
		grpc.UnaryInterceptor(insertFullMethod),
		grpc.StreamInterceptor(insertFullMethodStream),
	}
	s := &APIServer{
		server:            grpc.NewServer(sopts...),
		modelStoreAddr:    modelAddr,
		scheduler:         driver,
		instanceScheduler: instanceScheduler,
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
