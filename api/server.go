package api

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/model"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type APIOffer chan *RunRequest

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
	offerChan      APIOffer
}

func NewAPIServer(c APIOffer, modelAddr string) *APIServer {
	s := &APIServer{
		server:         grpc.NewServer(),
		offerChan:      c,
		modelStoreAddr: modelAddr,
	}
	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
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

type InstanceAPI struct {
	api *APIServer
}

func (s *InstanceAPI) Run(ctx context.Context, in *RunRequest) (*RunReply, error) {
	log.Printf("New Request: %v\n", in.String())
	// TODO: Rewrite using RPC filter mechanism.
	_, err := model.Connect([]string{s.api.modelStoreAddr})
	if err != nil {
		return nil, err
	}
	defer model.Close()
	inst, err := model.CreateInstance(&model.Instance{})
	s.api.offerChan <- in
	return &RunReply{InstanceId: inst.Id}, nil
}

type ResourceAPI struct {
	api *APIServer
}

func (s *ResourceAPI) Register(context.Context, *ResourceRequest) (*ResourceReply, error) {
	return &ResourceReply{}, nil
}
func (s *ResourceAPI) Unregister(context.Context, *ResourceIDRequest) (*ResourceReply, error) {
	return &ResourceReply{}, nil
}
