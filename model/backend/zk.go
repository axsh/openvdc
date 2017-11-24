package backend

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
)

const versionAny = int32(-1)

var defaultACL = zk.WorldACL(zk.PermAll)
var ErrFindLastKey = func(key string) error {
	return fmt.Errorf("Unable to find znode with last key: %s", key)
}

const DefaultBasePath = "/openvdc"

// Implements pflag.Value and backend.ConnectionAddress
type ZkEndpoint struct {
	Path  string
	Hosts []string // "host" or "host:port"
}

func (ZkEndpoint) isConnectionAddress() {}

func (ze *ZkEndpoint) String() string {
	return fmt.Sprintf("zk://%s%s", strings.Join(ze.Hosts, ","), ze.Path)
}

func (ze *ZkEndpoint) Set(value string) error {
	if strings.HasPrefix(value, "zk://") {
		zkurl, err := url.Parse(value)
		if err != nil {
			return errors.Wrap(err, "Invalid zk URL")
		}
		ze.Hosts = strings.Split(zkurl.Host, ",")
		ze.Path = zkurl.Path
	} else {
		host, port, err := net.SplitHostPort(value)
		if err != nil {
			host = value
			port = "2181"
		}
		if host == "" {
			host = "localhost"
		}
		ze.Hosts = []string{net.JoinHostPort(host, port)}
	}
	if ze.Path == "" {
		ze.Path = DefaultBasePath
	}
	return nil
}

func (ZkEndpoint) Type() string {
	return "ZkEndpoint"
}

type zkConnection struct {
	conn     *zk.Conn
	ev       <-chan zk.Event
	basePath string
	// TODO: Add mutex
}

func (z *zkConnection) Connect(dest ConnectionAddress) error {
	if z.isConnected() {
		return ErrConnectionExists
	}
	zkAddr, ok := dest.(ZkEndpoint)
	if !ok {
		p, ok := dest.(*ZkEndpoint)
		if !ok {
			return fmt.Errorf("Invalid connection address type: %T", dest)
		}
		zkAddr = *p
	}
	z.basePath = zkAddr.Path
	c, ev, err := zk.Connect(zkAddr.Hosts, 10*time.Second)
	if err != nil {
		return errors.Wrapf(err, "Failed zk.Connect")
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
		if ev.Type != zk.EventSession {
			// Ignore Node events.
			continue
		}
		if ev.State == zk.StateDisconnected {
			return nil
		}
		log.Warn("ZK disconnecting... ", ev.State)
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
	return &Zk{}
}

func (z *Zk) CreateWithID(key string, value []byte) (string, error) {
	if !z.isConnected() {
		return "", ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	nkey, err := z.connection().Create(absKey, value, zk.FlagSequence, defaultACL)

	// Dirty hack to work around zookeeper's connection closed error.
	retryCount := 0
	for err != nil && retryCount < 5 {
		sleep(2)
		retryCount++
		nkey, err = z.connection().Create(absKey, value, zk.FlagSequence, defaultACL)
	}
	if err != nil {
		return "", errors.Wrapf(err, "Failed zk.Create with sequence %s", absKey)
	}

	// Treat the parent key as the sequence store to save the last ID
	seqNode, nid := path.Split(nkey)
	seqNode = strings.TrimSuffix(seqNode, "/")
	_, err = z.connection().Set(seqNode, []byte(nid), versionAny)
	if err != nil {
		// Rollback by deleting new key node.
		z.connection().Delete(nkey, versionAny)
		return "", errors.Wrapf(err, "Failed updating last sequence ID zk.Set %s", seqNode)
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
		return errors.Wrapf(err, "Failed zk.Create %s", absKey)
	}
	return nil
}

func (z *Zk) Update(key string, value []byte) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	if _, err := z.connection().Set(absKey, value, versionAny); err != nil {
		return errors.Wrapf(err, "Failed zk.Set %s", absKey)
	}
	return nil
}

func (z *Zk) Find(key string) ([]byte, error) {
	if !z.isConnected() {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	_, err := z.connection().Sync(absKey)
	if err != nil {
		log.WithError(err).Warnf("Failed zk.Sync: %s", absKey)
	}
	value, _, err := z.connection().Get(absKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed zk.Get %s", absKey)
	}
	return value, nil
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
	return errors.Wrapf(z.connection().Delete(absKey, 1), "Failed zk.Delete %s", absKey)
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
		return nil, errors.Wrapf(err, "Failed zk.Children %s", absKey)
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
			return errors.Wrapf(err, "Failed zk.Exists %s", z.zk.basePath)
		}
		if !ok {
			// TODO: Set the current schema version to the root key node.
			// the data should be human readable.
			_, err = conn.Create(z.zk.basePath, []byte{}, int32(0), defaultACL)
			if err != nil {
				log.WithError(err).Error("Failed to setup the root key.")
				return errors.Wrapf(err, "Failed zk.Create %s", z.zk.basePath)
			}
		}
	}
	for _, key := range subkeys {
		absKey, _ := z.zk.canonKey(key)
		ok, _, err := conn.Exists(absKey)
		if err != nil {
			return errors.Wrapf(err, "Failed zk.Exists %s", absKey)
		}
		if !ok {
			_, err := conn.Create(absKey, []byte{}, int32(0), defaultACL)
			if err != nil {
				log.WithError(err).Errorf("Failed to setup the schema key: %s (%s)", key, absKey)
				return errors.Wrapf(err, "Failed zk.Create %s", absKey)
			}
		}
	}
	return nil
}

func (z *Zk) Watch(key string) (WatchEvent, error) {
	if !z.isConnected() {
		return EventErr, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	var err error
	var ev <-chan zk.Event
	if strings.HasSuffix(absKey, "/*") {
		_, _, ev, err = z.conn.ChildrenW(path.Dir(absKey))
		if err != nil {
			return EventErr, errors.Wrap(err, "conn.ChildrenW")
		}
	} else {
		var exists bool
		exists, _, ev, err = z.conn.ExistsW(absKey)
		if err != nil {
			return EventErr, errors.Wrap(err, "conn.ExistsW")
		}
		if !exists {
			return EventErr, errors.Errorf("Unknown key: %s", absKey)
		}
	}
	received := <-ev
	switch received.Type {
	case zk.EventNodeCreated:
		return EventCreated, nil
	case zk.EventNodeDeleted:
		return EventDeleted, nil
	case zk.EventNodeDataChanged:
		return EventModified, nil
	}

	return EventUnknown, nil
}
