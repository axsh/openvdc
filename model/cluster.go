package model

import (
	"fmt"
	"time"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
)

const clusterBaseKey = "cluster"

func init() {
	schemaKeys = append(schemaKeys, clusterBaseKey)
}

type ClusterOps interface {
	Register(*ClusterNode) error
}

type cluster struct {
	ctx context.Context
}

func Cluster(ctx context.Context) ClusterOps {
	return &cluster{ctx: ctx}
}

func (i *cluster) connection() (backend.ClusterBackend, error) {
	bk := GetClusterBackendCtx(i.ctx)
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	return bk, nil
}

func (i *cluster) Register(n *ClusterNode) error {
	if n.Id == "" {
		return fmt.Errorf("ID is not set")
	}

	createdAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	n.CreatedAt = createdAt
	data, err := proto.Marshal(n)
	if err != nil {
		return err
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	err = bk.Register(fmt.Sprintf("%s/%s", clusterBaseKey, n.Id), data)
	if err != nil {
		return err
	}
	return nil
}
