package model

import (
	"fmt"
	"path"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
)

type ResourceOps interface {
	Create(*Resource) (*Resource, error)
	FindByID(string) (*Resource, error)
}

type resources struct {
	ctx context.Context
}

func Resources(ctx context.Context) ResourceOps {
	return &resources{ctx: ctx}
}

func (i *resources) connection() (backend.ModelBackend, error) {
	bk := GetBackendCtx(i.ctx)
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	return bk, nil
}

func (i *resources) Create(n *Resource) (*Resource, error) {
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID("/resources/r-", data)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	return n, nil
}

func (i *resources) FindByID(id string) (*Resource, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	v, err := bk.Find(fmt.Sprintf("/resources/%s", id))
	if err != nil {
		return nil, err
	}
	n := &Resource{}
	err = proto.Unmarshal(v, n)
	if err != nil {
		return nil, err
	}
	n.Id = id
	return n, nil
}
