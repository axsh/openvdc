package api

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"golang.org/x/net/context"
)

var ErrTemplateUndefined = errors.New("Template is undefined")
var ErrUnknownTemplate = errors.New("Unknown template type")

type ResourceAPI struct {
	api *APIServer
}

func (s *ResourceAPI) Register(ctx context.Context, in *ResourceRequest) (*ResourceReply, error) {
	if in.GetTemplate() == nil {
		log.WithError(ErrTemplateUndefined).Error("template parameter is nil")
		return nil, ErrTemplateUndefined
	}
	r := &model.Resource{
		Template: in.GetTemplate(),
	}
	resource, err := model.Resources(ctx).Create(r)
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &ResourceReply{ID: resource.GetId(), Resource: resource}, nil
}
func (s *ResourceAPI) Unregister(ctx context.Context, in *ResourceIDRequest) (*ResourceReply, error) {
	// in.Key takes nil possibly.
	if in.GetKey() == nil {
		log.Error("Invalid resource identifier")
		return nil, fmt.Errorf("Invalid resource identifier")
	}
	// TODO: handle the case for in.GetName() is received.
	err := model.Resources(ctx).Destroy(in.GetID())
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	return &ResourceReply{ID: in.GetID()}, nil
}

func (s *ResourceAPI) Show(ctx context.Context, in *ResourceIDRequest) (*ResourceReply, error) {
	// in.Key takes nil possibly.
	if in.GetKey() == nil {
		log.Error("Invalid resource identifier")
		return nil, fmt.Errorf("Invalid resource identifier")
	}
	// TODO: handle the case for in.GetName() is received.
	resource, err := model.Resources(ctx).FindByID(in.GetID())
	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	return &ResourceReply{ID: resource.GetId(), Resource: resource}, nil
}
