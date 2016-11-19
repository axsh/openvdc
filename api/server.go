package api

import (
	"net"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/model"

	"github.com/axsh/openvdc/model/backend"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"
)

type APIOffer chan *RunRequest

type APIServer struct {
	server         *grpc.Server
	modelStoreAddr string
	offerChan      APIOffer
}

func NewAPIServer(c APIOffer, modelAddr string) *APIServer {
	sopts := []grpc.ServerOption{
		// Setup request middleware for the model.backend database connection.
		grpc.InTapHandle(func(ctx context.Context, info *tap.Info) (context.Context, error) {
			bk := backend.NewZkBackend()
			err := bk.Connect([]string{modelAddr})
			if err != nil {
				log.WithError(err).Errorf("Failed to connect to model backend: %s", modelAddr)
				return ctx, err
			}
			ctx = withBackendCtx(ctx, bk)
			go func() {
				<-ctx.Done()

				bk, ok := getBackendCtx(ctx)
				// Assert returned type from ctx.
				if !ok {
					log.Fatalf("Unexpected type to '%s' context value: %v", ctxBackendKey, reflect.TypeOf(bk))
				}
				err := bk.Close()
				if err != nil {
					log.WithError(err).Error("Failed to close connection to model backend.")
				}
			}()
			return ctx, nil
		}),
	}
	s := &APIServer{
		server:         grpc.NewServer(sopts...),
		offerChan:      c,
		modelStoreAddr: modelAddr,
	}
	RegisterInstanceServer(s.server, &InstanceAPI{api: s})
	RegisterResourceServer(s.server, &ResourceAPI{api: s})
	return s
}

type ctxKey string

const ctxBackendKey ctxKey = "model.backend"

func withBackendCtx(ctx context.Context, bk backend.ModelBackend) context.Context {
	return context.WithValue(ctx, ctxBackendKey, bk)
}

func getBackendCtx(ctx context.Context) (backend.ModelBackend, bool) {
	bk, ok := ctx.Value(ctxBackendKey).(backend.ModelBackend)
	return bk, ok
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
	log.Info("Called Register()")
	return &ResourceReply{ID: "r-00000001"}, nil
}
func (s *ResourceAPI) Unregister(context.Context, *ResourceIDRequest) (*ResourceReply, error) {
	return &ResourceReply{ID: "r-00000001"}, nil
}
