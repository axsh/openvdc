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

func GrpcInterceptor(modelAddr []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := Connect(ctx, modelAddr)
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
