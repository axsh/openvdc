package model

import (
	"fmt"

	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

const clusterBaseKey = "cluster"

func init() {
	schemaKeys = append(schemaKeys, clusterBaseKey)
}

// ClusterNode is a marker interface. Protobuf object implements this
// is allowed to store in /cluster key space.
type ClusterNode interface {
	proto.Message
	isClusterNode()
	GetId() string
}

func (ExecutorNode) isClusterNode()  {}
func (SchedulerNode) isClusterNode() {}
func (MonitorNode) isClusterNode()   {}

type ClusterOps interface {
	Register(node ClusterNode) error
	Find(nodeID string, node ClusterNode) error
	Unregister(nodeID string) error
}

type cluster struct {
	ctx context.Context
}

func Cluster(ctx context.Context) ClusterOps {
	return &cluster{ctx: ctx}
}

func (i *cluster) connection() (backend.ProtoClusterBackend, error) {
	bk := GetClusterBackendCtx(i.ctx)
	if bk == nil {
		return nil, ErrBackendNotInContext
	}
	wrapper := backend.NewProtoClusterWrapper(bk)
	return wrapper, nil
}

func (i *cluster) Register(n ClusterNode) error {
	if n.GetId() == "" {
		return fmt.Errorf("ID is not set")
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	if err := bk.Register(fmt.Sprintf("%s/%s", clusterBaseKey, n.GetId()), n); err != nil {
		return err
	}
	return nil
}

func (i *cluster) Find(nodeID string, in ClusterNode) error {
	if nodeID == "" {
		return fmt.Errorf("ID is not set")
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	if err := bk.Find(fmt.Sprintf("%s/%s", clusterBaseKey, nodeID), in); err != nil {
		return err
	}
	return nil
}

func (i *cluster) Unregister(nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("ID is not set")
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}
	if err := bk.Unregister(fmt.Sprintf("%s/%s", clusterBaseKey, nodeID)); err != nil {
		return err
	}
	return nil
}
