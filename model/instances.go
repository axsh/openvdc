package model

import (
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
