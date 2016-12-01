package backend

import (
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

func withConnect(t *testing.T, c func(z *Zk)) (err error) {
	z := NewZkBackend()
	z.basePath = "/" + uuid.New()
	defer func() {
		err = z.Delete(z.basePath)
		z.Close()
	}()
	err = z.Connect([]string{unittest.TestZkServer})
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

func TestZkCreateWithID(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{})
		assert.NoError(err)
		nkey, err := z.CreateWithID("/test1/t-", []byte{})
		assert.NoError(err)
		z.Delete(nkey)
		z.Delete("/test1")
	})
}

func TestZkCreate2AndFindLast(t *testing.T) {
	assert := assert.New(t)
	withConnect(t, func(z *Zk) {
		err := z.Create("/test1", []byte{})
		assert.NoError(err)
		nkey1, err := z.CreateWithID2("/test1/t-", []byte{})
		assert.NoError(err)
		nkey2, err := z.CreateWithID2("/test1/t-", []byte{})
		assert.NoError(err)
		lkey, err := z.FindLastKey("/test1/t-")
		assert.NoError(err)
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
		assert.Implements((*ModelSchema)(nil), z)
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
