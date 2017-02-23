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
	GetUUID() string
}

type NodeOps interface {
	Add(node Node) error
	Find(nodeUUID string, node Node) error
}

type nodes struct {
	base
}

func Nodes(ctx context.Context) NodeOps {
	return &nodes{base{ctx: ctx}}
}

func (i *nodes) Add(n Node) error {
	if n.GetUUID() == "" {
		return fmt.Errorf("ID is not set")
	}

	bk, err := i.connection()
	if err != nil {
		return err
	}

	if err = bk.Backend().Create(fmt.Sprintf("%s/%v", nodesBaseKey, n.GetUUID()), []byte{}); err != nil {
		return nil
	}

	return nil
}

func (i *nodes) Find(nodeUUID string, in Node) error {

	return nil
}
