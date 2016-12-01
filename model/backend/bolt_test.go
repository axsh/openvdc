package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoltBackend(t *testing.T) {
	assert := assert.New(t)
	z := NewBoltBackend()
	assert.Implements((*ModelBackend)(nil), z)
}

func TestBoltConnect(t *testing.T) {
	assert := assert.New(t)
	z := NewBoltBackend()
	defer z.Close()
	err := z.Connect([]string{"my.db"})
	assert.NoError(err)
}

func TestBoltCreate(t *testing.T) {
	assert := assert.New(t)
	z := NewBoltBackend()
	defer z.Close()
	err := z.Connect([]string{"my.db"})
	assert.NoError(err)
	err = z.Create("/test1", []byte{})
	assert.NoError(err)
}
