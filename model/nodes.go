package model

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

const nodesBaseKey = "nodes"

func init() {
	schemaKeys = append(schemaKeys, nodesBaseKey)
}

type Node interface {
	proto.Message
	GetAgentID() string
	GetAgentMesosID() string
}

type NodeOps interface {
	Add(node Node) error
	FindByAgentMesosID(agentMesosID string) (*AgentNode, error)
	FindByAgentID(agentID string) (*AgentNode, error)
	Filter(limit int, cb func(*AgentNode) int) error
	UpdateAgentMesosID(agentID string, agentMesosID string) error
}

type nodes struct {
	base
}

func Nodes(ctx context.Context) NodeOps {
	return &nodes{base{ctx: ctx}}
}

func (i *nodes) Add(n Node) error {
	if n.GetAgentID() == "" {
		return fmt.Errorf("ID is not set")
	}

	agentNode, _ := i.FindByAgentID(n.GetAgentID())

	if agentNode != nil {
		return nil
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}

	buf, err := proto.Marshal(n)
	if err != nil {
		return err
	}

	if err = bk.Backend().Create(fmt.Sprintf("%s/%s", nodesBaseKey, n.GetAgentID()), buf); err != nil {
		return nil
	}

	return nil
}

func (i *nodes) FindByAgentID(agentID string) (*AgentNode, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	n := &AgentNode{}
	if err := bk.Find(fmt.Sprintf("/%s/%s", nodesBaseKey, agentID), n); err != nil {
		return nil, err
	}

	return n, nil
}

func (i *nodes) FindByAgentMesosID(agentMesosID string) (*AgentNode, error) {
	res := []*AgentNode{}
	err := i.Filter(1, func(node *AgentNode) int {
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

func (i *nodes) Filter(limit int, cb func(*AgentNode) int) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}
	keys, err := bk.Keys(fmt.Sprintf("/%s", nodesBaseKey))
	if err != nil {
		return err
	}
	for keys.Next() {
		node, err := i.FindByAgentID(keys.Value())
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

func (i *nodes) UpdateAgentMesosID(agentID string, agentMesosID string) error {
	node, err := i.FindByAgentID(agentID)
	if err != nil {
		return err
	}

	node.Agentmesosid = agentMesosID

	bk, err := i.connection()
	if err != nil {
		return err
	}

	err = bk.Update(fmt.Sprintf("/%s/%s/", nodesBaseKey, node.Agentid), node)
	if err != nil {
		return err
	}
	return nil
}
