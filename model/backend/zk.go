package backend

import (
	"path"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/samuel/go-zookeeper/zk"
)

var defaultACL = zk.WorldACL(zk.PermAll)

type Zk struct {
	conn     *zk.Conn
	basePath string
}

func NewZkBackend() *Zk {
	return &Zk{
		basePath: "/openvdc",
	}
}

func (z *Zk) Connect(servers []string) error {
	if z.conn != nil {
		return ErrConnectionExists
	}
	c, f, err := zk.Connect(servers, time.Second)
	if err != nil {
		return err
	}
	<-f
	z.conn = c
	return nil
}

func (z *Zk) Close() error {
	if z.conn == nil {
		return ErrConnectionNotReady
	}
	defer func() {
		z.conn = nil
	}()
	z.conn.Close()
	return nil
}

func (z *Zk) ifNotExist(key string, callee func() error) error {
	exists, _, err := z.conn.Exists(key)
	if err != nil {
		return err
	}
	if !exists {
		return callee()
	}
	return nil
}

func (z *Zk) create(key string, value []byte, flags int32) (string, error) {
	if z.conn == nil {
		return "", ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	return z.conn.Create(absKey, value, flags, defaultACL)
}

func (z *Zk) CreateWithID(key string, value []byte) (string, error) {
	nkey, err := z.create(key, value, zk.FlagSequence)
	if err != nil {
		return "", err
	}
	return nkey, nil
}

func (z *Zk) Create(key string, value []byte) error {
	_, err := z.create(key, value, int32(0))
	return err
}

func (z *Zk) Update(key string, value []byte) error {
	if z.conn == nil {
		return ErrConnectionNotReady
	}
	var err error
	absKey, _ := z.canonKey(key)

	ok, stat, err := z.conn.Exists(absKey)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnknownKey(absKey)
	}
	_, err = z.conn.Set(absKey, value, stat.Version)
	return err
}

func (z *Zk) Find(key string) (value []byte, err error) {
	if z.conn == nil {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	value, _, err = z.conn.Get(absKey)
	return
}

func (z *Zk) canonKey(key string) (absKey string, dir string) {
	absKey = path.Clean(path.Join(z.basePath, key))
	dir, _ = path.Split(absKey)
	dir = strings.TrimSuffix(dir, "/")
	return absKey, dir
}

func (z *Zk) Delete(key string) error {
	if z.conn == nil {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	return z.conn.Delete(absKey, 1)
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
	if z.conn == nil {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(parentKey)
	children, _, err := z.conn.Children(absKey)
	if err != nil {
		return nil, err
	}
	// Sort keys in client side.
	sort.Strings(children)
	return &childIt{children: &children, cur: -1}, nil
}

func (z *Zk) Schema() SchemaHandler {
	return &zkSchemaHandler{zk: z}
}

type zkSchemaHandler struct {
	zk *Zk
}

func (z *zkSchemaHandler) Install(subkeys []string) error {
	conn := z.zk.conn
	if conn == nil {
		return ErrConnectionNotReady
	}
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
