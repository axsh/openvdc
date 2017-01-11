package model

import (
	"time"

	"github.com/axsh/openvdc/model/backend"
	"golang.org/x/net/context"
)

type base struct {
	ctx context.Context
}

func (i *base) connection() (backend.ProtoModelBackend, error) {
	bk := backend.NewProtoWrapper(GetBackendCtx(i.ctx))
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	bk.AddFilter(&backend.TimestampFilter{time.Now().UTC()})
	return bk, nil
}
