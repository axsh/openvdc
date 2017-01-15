package backend

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/samuel/go-zookeeper/zk"
)

const versionAny = int32(-1)

var defaultACL = zk.WorldACL(zk.PermAll)
var ErrFindLastKey = func(key string) error {
	return fmt.Errorf("Unable to find znode with last key: %s", key)
}

type zkConnection struct {
	conn     *zk.Conn
	ev       <-chan zk.Event
	basePath string
}

func (z *zkConnection) Connect(servers []string) error {
	if z.isConnected() {
		return ErrConnectionExists
	}
	c, ev, err := zk.Connect(servers, 10*time.Second)
	if err != nil {
		return err
	}

	for e := range ev {
		switch e.State {
		case zk.StateHasSession:
			// Set members if connected successfully.
			z.conn = c
			z.ev = ev
			return nil
		case zk.StateConnecting, zk.StateConnected, zk.StateConnectedReadOnly:
			// Pass
		default:
			log.Errorf(e.State.String())
		}
	}

	return nil
}

func (z *zkConnection) Close() error {
	if z.conn == nil {
		return ErrConnectionNotReady
	}
	defer func() {
		z.conn = nil
		z.ev = nil
	}()
	z.conn.Close()
	for ev := range z.ev {
		if ev.State == zk.StateDisconnected {
			return nil
		} else {
			log.Warn("ZK disconnecting... ", ev.State.String())
		}
	}
	return nil
}

func (z *zkConnection) connection() *zk.Conn {
	return z.conn
}

func (z *zkConnection) isConnected() bool {
	if z.conn == nil {
		return false
	}
	return z.conn.State() == zk.StateHasSession
}

func (z *zkConnection) canonKey(key string) (absKey string, dir string) {
	absKey = path.Clean(path.Join(z.basePath, key))
	dir, _ = path.Split(absKey)
	dir = strings.TrimSuffix(dir, "/")
	return absKey, dir
}

func (z *zkConnection) Schema() SchemaHandler {
	return &zkSchemaHandler{zk: z}
}

type Zk struct {
	zkConnection
	// TODO: Add mutex
}

func NewZkBackend() *Zk {
	return &Zk{
		zkConnection: zkConnection{
			basePath: "/openvdc",
		},
	}
}

func (z *Zk) CreateWithID(key string, value []byte) (string, error) {
	if !z.isConnected() {
		return "", ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	nkey, err := z.connection().Create(absKey, value, zk.FlagSequence, defaultACL)
	if err != nil {
		return "", err
	}

	// Treat the parent key as the sequence store to save the last ID
	seqNode, nid := path.Split(nkey)
	seqNode = strings.TrimSuffix(seqNode, "/")
	_, err = z.connection().Set(seqNode, []byte(nid), versionAny)
	if err != nil {
		// Rollback by deleting new key node.
		z.connection().Delete(nkey, versionAny)
		return "", err
	}
	return nkey[len(z.basePath):], nil
}

func (z *Zk) Create(key string, value []byte) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	_, err := z.connection().Create(absKey, value, int32(0), defaultACL)
	if err != nil {
		return err
	}
	return nil
}

func (z *Zk) Update(key string, value []byte) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	var err error
	absKey, _ := z.canonKey(key)

	ok, stat, err := z.connection().Exists(absKey)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnknownKey(absKey)
	}
	_, err = z.connection().Set(absKey, value, stat.Version)
	return err
}

func (z *Zk) Find(key string) (value []byte, err error) {
	if !z.isConnected() {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	value, _, err = z.connection().Get(absKey)
	return
}

func (z *Zk) FindLastKey(prefixKey string) (string, error) {
	if !z.isConnected() {
		return "", ErrConnectionNotReady
	}
	_, base := z.canonKey(prefixKey)
	buf, _, err := z.connection().Get(base)
	if err != nil {
		return "", ErrFindLastKey(err.Error())
	}
	if len(buf) == 0 {
		return "", ErrFindLastKey("<empty>")
	}
	lastKey := path.Join(base, string(buf))
	return lastKey[len(z.basePath):], nil
}

func (z *Zk) Delete(key string) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	return z.connection().Delete(absKey, 1)
}

// TODO: Add mutex for thread safety.
type childIt struct {
	cur      int
	children *[]string
}

func (c *childIt) Next() bool {
	if len(*c.children) > c.cur {
		c.cur++
	}
	return (len(*c.children) > c.cur)
}

func (c *childIt) Value() string {
	if c.cur < 0 || len(*c.children) == 0 {
		return ""
	} else if len(*c.children) == c.cur {
		return (*c.children)[c.cur-1]
	}
	return (*c.children)[c.cur]
}

func (z *Zk) Keys(parentKey string) (KeyIterator, error) {
	if !z.isConnected() {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(parentKey)
	children, _, err := z.connection().Children(absKey)
	if err != nil {
		return nil, err
	}
	// Sort keys in client side.
	sort.Strings(children)
	return &childIt{children: &children, cur: -1}, nil
}

type zkSchemaHandler struct {
	zk *zkConnection
}

func (z *zkSchemaHandler) Install(subkeys []string) error {
	if !z.zk.isConnected() {
		return ErrConnectionNotReady
	}
	conn := z.zk.conn
	{
		// Install the root key if not exists.
		ok, _, err := conn.Exists(z.zk.basePath)
		if err != nil {
			return err
		}
		if !ok {
			// TODO: Set the current schema version to the root key node.
			// the data should be human readable.
			_, err = conn.Create(z.zk.basePath, []byte{}, int32(0), defaultACL)
			if err != nil {
				log.WithError(err).Error("Failed to setup the root key.")
				return err
			}
		}
	}
	for _, key := range subkeys {
		absKey, _ := z.zk.canonKey(key)
		ok, _, err := conn.Exists(absKey)
		if err != nil {
			return err
		}
		if !ok {
			_, err := conn.Create(absKey, []byte{}, int32(0), defaultACL)
			if err != nil {
				log.WithError(err).Errorf("Failed to setup the schema key: %s (%s)", key, absKey)
				return err
			}
		}
	}
	return nil
}
