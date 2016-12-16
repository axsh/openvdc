package backend

import "errors"

var ErrConnectionNotReady = errors.New("Connection is not established yet.")
var ErrConnectionExists = errors.New("Connection is established")
var ErrUnknownKey = func(key string) error {
	return errors.New("Unknown key name: " + key)
}

type KeyIterator interface {
	Next() bool
	Value() string
}

type ModelBackend interface {
	Connect(dest []string) error
	Close() error
	Create(key string, value []byte) error
	CreateWithID(key string, value []byte) (string, error)
	Update(key string, value []byte) error
	Find(key string) ([]byte, error)
	Delete(key string) error
	Keys(parentKey string) (KeyIterator, error)
	FindLastKey(prefixKey string) (string, error)
}

type ModelSchema interface {
	Schema() SchemaHandler
}

type SchemaHandler interface {
	Install(subkeys []string) error
}

type ClusterBackend interface {
	Connect(dest []string) error
	Close() error
	Register(key string, value []byte) error
}
