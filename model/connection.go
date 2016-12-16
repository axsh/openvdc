package model

import (
	"errors"
	"reflect"
	"runtime"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model/backend"
	"golang.org/x/net/context"
)

var ErrBackendNotInContext = errors.New("Given context does not have the backend object")

func Connect(ctx context.Context, dest []string) (context.Context, error) {
	bk := backend.NewZkBackend()
	err := bk.Connect(dest)
	if err != nil {
		return nil, err
	}
	return withBackendCtx(ctx, bk), nil
}

func Close(ctx context.Context) error {
	bk := GetBackendCtx(ctx)
	if bk == nil {
		return ErrBackendNotInContext
	}
	return bk.Close()
}

var schemaKeys []string

func InstallSchemas(bk backend.ModelSchema) error {
	schema := bk.Schema()
	return schema.Install(schemaKeys)
}

type ctxKey string

const ctxBackendKey ctxKey = "model.backend"

func withBackendCtx(ctx context.Context, bk backend.ModelBackend) context.Context {
	return context.WithValue(ctx, ctxBackendKey, bk)
}

func GetBackendCtx(ctx context.Context) backend.ModelBackend {
	if ctx == nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("GetBackendCtx() does not accept nil.: %s:%d", file, line)
	}
	bk, ok := ctx.Value(ctxBackendKey).(backend.ModelBackend)
	// Assert returned type from ctx.
	if !ok && bk != nil {
		log.Fatalf("Unexpected type to '%s' context value: %v", ctxBackendKey, reflect.TypeOf(bk))
	}
	return bk
}

func GrpcInterceptor(modelAddr string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := Connect(ctx, []string{modelAddr})
		if err != nil {
			log.WithError(err).Errorf("Failed to connect to model backend: %s", modelAddr)
			return nil, err
		}
		defer func() {
			err := Close(ctx)
			if err != nil {
				log.WithError(err).Error("Failed to close connection to model backend.")
			}
		}()
		return handler(ctx, req)
	}
}

func GrpcStreamInterceptor(modelAddr string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := Connect(ss.Context(), []string{modelAddr})
		if err != nil {
			log.WithError(err).Errorf("Failed to connect to model backend: %s", modelAddr)
			return err
		}
		defer func() {
			err := Close(ctx)
			if err != nil {
				log.WithError(err).Error("Failed to close connection to model backend.")
			}
		}()
		return handler(srv, ss)
	}
}

const ctxClusterBackendKey ctxKey = "cluster.backend"

func withClusterBackendCtx(ctx context.Context, bk backend.ClusterBackend) context.Context {
	return context.WithValue(ctx, ctxClusterBackendKey, bk)
}

func ClusterConnect(ctx context.Context, dest []string) (context.Context, error) {
	bk := backend.NewZkClusterBackend()
	err := bk.Connect(dest)
	if err != nil {
		return nil, err
	}
	return withClusterBackendCtx(ctx, bk), nil
}

func ClusterClose(ctx context.Context) error {
	bk := GetClusterBackendCtx(ctx)
	if bk == nil {
		return ErrBackendNotInContext
	}
	return bk.Close()
}

func GetClusterBackendCtx(ctx context.Context) backend.ClusterBackend {
	if ctx == nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("BUGON: GetClusterBackendCtx() does not accept nil.: %s:%d", file, line)
	}
	bk, ok := ctx.Value(ctxClusterBackendKey).(backend.ClusterBackend)
	// Assert returned type from ctx.
	if !ok && bk != nil {
		log.Fatalf("BUGON: Unexpected type to '%s' context value: %v", ctxClusterBackendKey, reflect.TypeOf(bk))
	}
	return bk
}
