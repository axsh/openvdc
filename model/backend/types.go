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

type ConnectionAddress interface {
	isConnectionAddress()
}

type ModelBackend interface {
	BackendConnection
	Create(key string, value []byte) error
	CreateWithID(key string, value []byte) (string, error)
	Update(key string, value []byte) error
	Find(key string) ([]byte, error)
	Delete(key string) error
	Keys(parentKey string) (KeyIterator, error)
	FindLastKey(prefixKey string) (string, error)
}

type WatchEvent int

const (
	EventErr WatchEvent = iota
	EventUnknown
	EventCreated
	EventDeleted
	EventModified
)

var eventToName = map[WatchEvent]string{
	EventErr:      "Error",
	EventUnknown:  "Unknown",
	EventCreated:  "Created",
	EventDeleted:  "Deleted",
	EventModified: "Modified",
}

func (e WatchEvent) String() string {
	return eventToName[e]
}

type ModelWatcher interface {
	Watch(key string) (WatchEvent, error)
}

type ModelSchema interface {
	Schema() SchemaHandler
}

type SchemaHandler interface {
	Install(subkeys []string) error
}

type ClusterBackend interface {
	BackendConnection
	Register(nodeID string, value []byte) error
	Find(nodeID string) ([]byte, error)
	UnRegister(nodeID string) error
}

type BackendConnection interface {
	Connect(dest ConnectionAddress) error
	Close() error
}
