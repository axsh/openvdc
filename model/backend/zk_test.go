package backend

import (
	"testing"

	"strings"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func withConnect(t *testing.T, c func(z *Zk)) (err error) {
	ze := &ZkEndpoint{}
	if err := ze.Set(unittest.TestZkServer); err != nil {
		t.Fatal("Invalid zookeeper address:", unittest.TestZkServer)
	}
	ze.Path = "/" + uuid.New()

	z := NewZkBackend()
	defer func() {
		err = z.Delete(z.basePath)
		z.Close()
	}()
	err = z.Connect(ze)
	if err != err {
		t.Fatal(err)
	}
	z.Schema().Install([]string{})
	c(z)
	return
}

func TestZkEndpoint(t *testing.T) {
	assert := assert.New(t)

	ze := &ZkEndpoint{}
	assert.Implements((*ConnectionAddress)(nil), ze)
	assert.Implements((*pflag.Value)(nil), ze)
}

func TestNewZkBackend(t *testing.T) {
	assert := assert.New(t)

	z := NewZkBackend()
	assert.Implements((*BackendConnection)(nil), z)
	assert.Implements((*ModelBackend)(nil), z)
	assert.Implements((*ModelSchema)(nil), z)
}

func TestNewZkClusterBackend(t *testing.T) {
	assert := assert.New(t)

	z := NewZkClusterBackend()
	assert.Implements((*BackendConnection)(nil), z)
	assert.Implements((*ClusterBackend)(nil), z)
	assert.Implements((*ModelSchema)(nil), z)
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

func TestZkCreateWithID(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{})
		assert.NoError(err)
		nkey, err := z.CreateWithID("/test1/t-", []byte{})
		assert.NoError(err)
		assert.True(strings.HasPrefix(nkey, "/test1/t-"))
		z.Delete(nkey)
		z.Delete("/test1")
	})
}

func TestZkCreateWithIDAndFindLast(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{})
		assert.NoError(err)
		nkey1, err := z.CreateWithID("/test1/t-", []byte{})
		assert.NoError(err)
		assert.True(strings.HasPrefix(nkey1, "/test1/t-"))
		nkey2, err := z.CreateWithID("/test1/t-", []byte{})
		assert.NoError(err)
		assert.True(strings.HasPrefix(nkey2, "/test1/t-"))
		lkey, err := z.FindLastKey("/test1/t-")
		assert.NoError(err)
		assert.True(strings.HasPrefix(lkey, "/test1/t-"))
		assert.NotEqual(nkey1, lkey)
		assert.Equal(nkey2, lkey)
		z.Delete(nkey1)
		z.Delete(nkey2)
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
		ms := z.Schema()
		err := ms.Install([]string{"subkey1", "subkey2"})
		assert.NoError(err)
		err = ms.Install([]string{"subkey1", "subkey2"})
		assert.NoError(err)
	})
}

func TestZkKeys(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		// Iterate empty keys.
		it, err := z.Keys("/")
		assert.NoError(err)
		assert.False(it.Next())
		assert.Equal("", it.Value())

		// Iterate exisiting keys.
		err = z.Create("/test1", []byte{1})
		assert.NoError(err)
		err = z.Create("/test2", []byte{1})
		assert.NoError(err)
		it, err = z.Keys("/")
		assert.NoError(err)
		assert.True(it.Next())
		assert.Equal("test1", it.Value())
		assert.True(it.Next())
		assert.Equal("test2", it.Value())
		assert.False(it.Next())
		assert.Equal("test2", it.Value())
		z.Delete("/test1")
		z.Delete("/test2")
	})
}
