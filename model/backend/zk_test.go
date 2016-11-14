package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZkConnect(t *testing.T) {
	assert := assert.New(t)
	z := NewZkBackend()
	defer z.Close()
	err := z.Connect([]string{"192.168.56.120"})
	assert.NoError(err)
}

func TestZkCreate(t *testing.T) {
	assert := assert.New(t)
	z := NewZkBackend()
	defer z.Close()
	err := z.Connect([]string{"192.168.56.120"})
	assert.NoError(err)
	err = z.Create("/test1", []byte{})
	assert.NoError(err)
}
