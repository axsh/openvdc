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
		err = z.Delete(z.basePath)
		z.Close()
	}()
	err = z.Connect([]string{testZkServer})
	if err != err {
		t.Fatal(err)
	}
	z.Schema().Install([]string{})
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
		z.Delete("/test1")
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
		z.Delete("/test1")
	})
}

func TestZkSchema(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		assert.Implements((*ModelSchema)(nil), z)
		ms := z.Schema()
		err := ms.Install([]string{"subkey1", "subkey2"})
		assert.NoError(err)
		err = ms.Install([]string{"subkey1", "subkey2"})
		assert.NoError(err)
	})
}
