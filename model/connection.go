package model

import "github.com/axsh/openvdc/model/backend"

var bk backend.ModelBackend

func Connect(dest []string) (backend.ModelBackend, error) {
	bk = backend.NewZkBackend()
	err := bk.Connect(dest)
	if err != nil {
		return nil, err
	}
	return bk, nil
}

func Close() error {
	if bk == nil {
		return backend.ErrConnectionNotReady
	}
	err := bk.Close()
	bk = nil
	return err
}
