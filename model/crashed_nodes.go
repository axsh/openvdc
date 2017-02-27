package model

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

const crashedNodesBaseKey = "crashed_nodes"

func init() {
	schemaKeys = append(schemaKeys, crashedNodesBaseKey)
}

type CrashedAgentNode interface {
	proto.Message
	GetReconnected() bool
	GetAgentID() string
}

type CrashedNodesOps interface {
	Add(node CrashedAgentNode) error
	Find(agentID string, node CrashedAgentNode) error
	SetReconnected(agentID string, node CrashedAgentNode) error
}

type crashedNodes struct {
	base
}

func CrashedNodes(ctx context.Context) CrashedNodesOps {
	return &crashedNodes{base{ctx: ctx}}
}

func (i *crashedNodes) Add(n CrashedAgentNode) error {

	if n.GetAgentID() == "" {
		return fmt.Errorf("ID is not set")
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}

	buf, err := proto.Marshal(n)
	if err != nil {
		return err
	}

	if err = bk.Backend().Create(fmt.Sprintf("%s/%v", crashedNodesBaseKey, uuid.New()), buf); err != nil {
		return nil
	}

	return nil
}

func (i *crashedNodes) Find(agentID string, in CrashedAgentNode) error {
	return nil
}

func (i *crashedNodes) SetReconnected(agentID string, in CrashedAgentNode) error {
	return nil
}
