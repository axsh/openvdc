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
	Destroy(string) error
}

const resourcesBaseKey = "resources"

func init() {
	schemaKeys = append(schemaKeys, resourcesBaseKey)
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
	n.State = Resource_REGISTERED
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID(fmt.Sprintf("/%s/r-", resourcesBaseKey), data)
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
	v, err := bk.Find(fmt.Sprintf("/%s/%s", resourcesBaseKey, id))
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

func (i *resources) Destroy(id string) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}
	n, err := i.FindByID(id)
	if err != nil {
		return err
	}
	if err := n.validateStateTransition(Resource_UNREGISTERED); err != nil {
		return err
	}
	n.State = Resource_UNREGISTERED
	data, err := proto.Marshal(n)
	if err != nil {
		return err
	}
	return bk.Update(fmt.Sprintf("/%s/%s", resourcesBaseKey, id), data)
}

func (r *Resource) validateStateTransition(next Resource_State) error {
	var result bool
	switch r.GetState() {
	case Resource_REGISTERED:
		result = (next == Resource_UNREGISTERED)
	}

	if result {
		return nil
	}

	return fmt.Errorf("Invalid state transition: %s -> %s",
		r.GetState().String(),
		next.String(),
	)
}
