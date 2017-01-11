package model

import (
	"context"
	"time"

	"github.com/axsh/openvdc/model/backend"
)

type base struct {
	ctx context.Context
}

func (i *base) connection() (backend.ProtoModelBackend, error) {
	bk := backend.NewProtoWrapper(GetBackendCtx(i.ctx))
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	bk.AddFilter(&backend.TimestampFilter{time.Now()})
	return bk, nil
}
