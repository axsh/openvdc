package model

import (
	"fmt"
	"path"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
)

func CreateInstance(n *Instance) (*Instance, error) {
	if bk == nil {
		return nil, backend.ErrConnectionNotReady
	}
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID("/instances/i-", data)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	return n, nil
}

func FindInstanceByID(id string) (*Instance, error) {
	if bk == nil {
		return nil, backend.ErrConnectionNotReady
	}
	v, err := bk.Find(fmt.Sprintf("/instances/%s", id))
	if err != nil {
		return nil, err
	}
	n := &Instance{}
	err = proto.Unmarshal(v, n)
	if err != nil {
		return nil, err
	}
	n.Id = id
	return n, nil
}
