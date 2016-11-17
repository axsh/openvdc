package backend

import (
	"os"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

var testZkServer string

func init() {
	if os.Getenv("ZK") != "" {
		testZkServer = os.Getenv("ZK")
	} else {
		testZkServer = "127.0.0.1"
	}
}

func withConnect(t *testing.T, c func(z *Zk)) (err error) {
	z := NewZkBackend()
	z.basePath = "/" + uuid.New()
	defer func() {
		err = z.conn.Delete(z.basePath, 1)
		z.Close()
	}()
	err = z.Connect([]string{testZkServer})
	if err != err {
		t.Fatal(err)
	}
	c(z)
	return
}

func TestZkConnect(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		assert.NotNil(z)
		assert.NotNil(z.conn)
	})
}

func TestZkCreate(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{})
		assert.NoError(err)
		err = z.Create("/test1", []byte{})
		assert.Error(err)
		// TODO: implemente ModelBackend.Delete().
		z.conn.Delete(z.basePath+"/test1", 1)
	})
}

func TestZkFind(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{1})
		assert.NoError(err)
		v, err := z.Find("/test1")
		assert.NoError(err)
		assert.Equal([]byte{1}, v)
		// TODO: implemente ModelBackend.Delete().
		z.conn.Delete(z.basePath+"/test1", 1)
	})
}
