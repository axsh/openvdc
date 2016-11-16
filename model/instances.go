package model

import (
	"os"
	"path"

	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
)

func connect() (backend.ModelBackend, error) {
	bk := backend.NewZkBackend()
	err := bk.Connect([]string{os.Getenv("ZK")})
	if err != nil {
		return nil, err
	}
	return bk, nil
}

func CreateInstance(n *Instance) (*Instance, error) {
	bk, err := connect()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(n)
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID("/openvdc/instances/i-", data)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	return n, nil
}
