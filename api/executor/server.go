package executor

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type ExecutorAPIServer struct {
	server         *grpc.Server
	listener       net.Listener
	modelStoreAddr string
}

func NewExecutorAPIServer(modelAddr string, ctx context.Context) *ExecutorAPIServer {
	// Assert the ctx has "cluster.backend" key
	model.GetClusterBackendCtx(ctx)

	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.UnaryInterceptor(model.GrpcInterceptor(modelAddr, ctx)),
		grpc.StreamInterceptor(model.GrpcStreamInterceptor(modelAddr, ctx)),
	}
	s := &ExecutorAPIServer{
		server:         grpc.NewServer(sopts...),
		modelStoreAddr: modelAddr,
	}

	RegisterInstanceConsoleServer(s.server, &InstanceConsoleAPI{api: s})
	return s
}

func (s *ExecutorAPIServer) Serve(listen net.Listener) error {
	s.listener = listen
	return s.server.Serve(listen)
}

func (s *ExecutorAPIServer) Stop() {
	s.server.Stop()
	s.listener = nil
}

func (s *ExecutorAPIServer) GracefulStop() {
	s.server.GracefulStop()
	s.listener = nil
}

func (s *ExecutorAPIServer) Listener() net.Listener {
	return s.listener
}

type InstanceConsoleAPI struct {
	api *ExecutorAPIServer
}

func (s *InstanceConsoleAPI) Attach(stream InstanceConsole_AttachServer) error {
	cin, err := stream.Recv()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Println(cin.InstanceId)
	return nil
}
