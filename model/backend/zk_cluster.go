package backend

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
)

type ZkCluster struct {
	conn     *zk.Conn
	basePath string
	ev       <-chan zk.Event
	// TODO: Add mutex
}

func NewZkClusterBackend() *ZkCluster {
	return &ZkCluster{
		basePath: "/openvdc",
	}
}

func (z *ZkCluster) Connect(servers []string) error {
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

func (z *ZkCluster) Close() error {
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

func (z *ZkCluster) isConnected() bool {
	if z.conn == nil {
		return false
	}
	return z.conn.State() == zk.StateHasSession
}

func (z *ZkCluster) canonKey(key string) (absKey string, dir string) {
	absKey = path.Clean(path.Join(z.basePath, key))
	dir, _ = path.Split(absKey)
	dir = strings.TrimSuffix(dir, "/")
	return absKey, dir
}

func (z *ZkCluster) Register(key string, value []byte) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)

	var errRetry = fmt.Errorf("")

	doRegist := func() error {
		_, err := z.conn.Create(absKey, value, zk.FlagEphemeral, defaultACL)
		if err != nil {
			if err == zk.ErrNodeExists {
				return errRetry
			}
			return err
		}
		return nil
	}

	for i := 0; i < 3; i++ {
		err := doRegist()
		if err == nil {
			return nil
		} else if err != errRetry {
			return err
		}
		log.Warnf("Retrying registration (%d)", i+1)
		time.Sleep(100 * time.Millisecond)
	}
	log.Errorf("Retry exceede.")
	return fmt.Errorf("Retry exceeded")
}

func (z *ZkCluster) Find(key string) (value []byte, err error) {
	if !z.isConnected() {
		return nil, ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)
	value, _, err = z.conn.Get(absKey)
	return
}
