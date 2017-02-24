package model

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

const crashedNodesBaseKey = "crashed_nodes"

func init() {
	schemaKeys = append(schemaKeys, crashedNodesBaseKey)
}

type CrashedNode interface {
	proto.Message
	GetUUID() string
	GetReconnected() string
}

type CrashedNodesOps interface {
	Add(node CrashedNode) error
	Find(nodeUUID string, node CrashedNode) error
	SetReconnected(nodeUUID string, node CrashedNode) error
}

type crashedNodes struct {
	base
}

func CrashedNodes(ctx context.Context) CrashedNodesOps {
	return &crashedNodes{base{ctx: ctx}}
}

func (i *crashedNodes) Add(n CrashedNode) error {
	return nil
}

func (i *crashedNodes) Find(nodeUUID string, in CrashedNode) error {
	return nil
}

func (i *crashedNodes) SetReconnected(nodeUUID string, in CrashedNode) error {
	return nil
}
