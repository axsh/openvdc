package backend

import "errors"

var ErrConnectionNotReady = errors.New("Connection is not established yet.")
var ErrConnectionExists = errors.New("Connection is established")

type ModelBackend interface {
	Connect(dest []string) error
	Close() error
	Create(key string, value []byte) error
	Update(key string, value []byte) error
}
