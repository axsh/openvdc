package model

import (
	"time"

	"github.com/axsh/openvdc/model/backend"
	"golang.org/x/net/context"
)

// TODO: move cluster.proto
//go:generate protoc -I../proto -I${GOPATH}/src --go_out=${GOPATH}/src ../proto/model.proto ../proto/cluster.proto ../proto/resource.proto

type base struct {
	ctx context.Context
}

func (i *base) connection() (backend.ProtoModelBackend, error) {
	bk := GetBackendCtx(i.ctx)
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	wrapper := backend.NewProtoWrapper(bk)
	wrapper.AddFilter(&backend.TimestampFilter{time.Now().UTC()})
	return wrapper, nil
}
