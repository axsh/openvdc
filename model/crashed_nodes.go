package model

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"time"
)

const crashedNodesBaseKey = "crashed_nodes"

func init() {
	schemaKeys = append(schemaKeys, crashedNodesBaseKey)
}

type CrashedAgentNode interface {
	proto.Message
	GetReconnected() bool
	GetAgentID() string
	GetUUID() string
}

type CrashedNodesOps interface {
	Add(node *CrashedNode) error
	FindByAgentMesosID(agentMesosID string) (*CrashedNode, error)
	FindByAgentID(agentID string) (*CrashedNode, error)
	FindByUUID(nodeUUID string) (*CrashedNode, error)
	Filter(limit int, cb func(*CrashedNode) int) error
	SetReconnected(node *CrashedNode) error
}

type crashedNodes struct {
	base
}

func CrashedNodes(ctx context.Context) CrashedNodesOps {
	return &crashedNodes{base{ctx: ctx}}
}

func (i *crashedNodes) Add(n *CrashedNode) error {
	if n.GetAgentID() == "" {
		return fmt.Errorf("ID is not set")
	}

	createdAt, _ := ptypes.TimestampProto(time.Now())

	n.Uuid = uuid.New()
	n.CreatedAt = createdAt

	bk, err := i.connection()
	if err != nil {
		return err
	}
	buf, err := proto.Marshal(n)
	if err != nil {
		return err
	}

	if err = bk.Backend().Create(fmt.Sprintf("%s/%v", crashedNodesBaseKey, n.Uuid), buf); err != nil {
		return nil
	}
	return nil
}

func (i *crashedNodes) Find(agentID string, in CrashedAgentNode) error {
	return nil
}

func (i *crashedNodes) SetReconnected(node *CrashedNode) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}

	reconnectedAt, _ := ptypes.TimestampProto(time.Now())

	updatedNode := &CrashedNode{
		Uuid:          node.GetUUID(),
		Agentid:       node.GetAgentID(),
		Agentmesosid:  node.GetAgentMesosID(),
		Reconnected:   true,
		CreatedAt:     node.GetCreatedAt(),
		ReconnectedAt: reconnectedAt,
	}
	return bk.Update(fmt.Sprintf("/%s/%s", crashedNodesBaseKey, node.Uuid), updatedNode)
}

func (i *crashedNodes) FindByAgentMesosID(agentMesosID string) (*CrashedNode, error) {
	res := []*CrashedNode{}
	err := i.Filter(1, func(node *CrashedNode) int {
		if node.GetAgentMesosID() == agentMesosID {
			res = append(res, node)
		}
		return len(res)
	})
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	} else {
		return nil, nil
	}
}

func (i *crashedNodes) FindByAgentID(agentID string) (*CrashedNode, error) {
	res := []*CrashedNode{}
	err := i.Filter(1, func(node *CrashedNode) int {
		if node.GetAgentID() == agentID && node.GetReconnected() == false {
			res = append(res, node)
		}
		return len(res)
	})
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0], nil
	} else {
		return nil, nil
	}
}

func (i *crashedNodes) Filter(limit int, cb func(*CrashedNode) int) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}
	keys, err := bk.Keys(fmt.Sprintf("/%s", crashedNodesBaseKey))
	if err != nil {
		return err
	}
	for keys.Next() {
		node, err := i.FindByUUID(keys.Value())
		if err != nil {
			return err
		}
		if limit > 0 && cb(node) > limit {
			break
		} else {
			cb(node)
		}
	}
	return nil
}

func (i *crashedNodes) FindByUUID(nodeUUID string) (*CrashedNode, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	n := &CrashedNode{}
	if err := bk.Find(fmt.Sprintf("/%s/%s", crashedNodesBaseKey, nodeUUID), n); err != nil {
		return nil, err
	}
	return n, nil
}
