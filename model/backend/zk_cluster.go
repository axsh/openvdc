package backend

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
)

type ZkCluster struct {
	zkConnection
	// TODO: Add mutex
}

func NewZkClusterBackend() *ZkCluster {
	return &ZkCluster{
		zkConnection: zkConnection{
			basePath: "/openvdc",
		},
	}
}

func (z *ZkCluster) Register(key string, value []byte) error {
	if !z.isConnected() {
		return ErrConnectionNotReady
	}
	absKey, _ := z.canonKey(key)

	var errRetry = fmt.Errorf("")

	doRegist := func() error {
		_, err := z.connection().Create(absKey, value, zk.FlagEphemeral, defaultACL)
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
	value, _, err = z.connection().Get(absKey)
	return
}
