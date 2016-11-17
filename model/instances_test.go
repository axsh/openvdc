package model

import (
	"os"
	"testing"

	"github.com/axsh/openvdc/model/backend"
	"github.com/stretchr/testify/assert"
)

func withConnect(t *testing.T, c func()) (err error) {
	_, err = Connect([]string{os.Getenv("ZK")})
	if err != nil {
		t.Fatal(err)
	}
	defer Close()
	c()
	return
}

func TestCreateInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		ExecutorId: "xxx",
	}
	_, err := CreateInstance(n)
	assert.Equal(backend.ErrConnectionNotReady, err)

	withConnect(t, func() {
		got, err := CreateInstance(n)
		assert.NoError(err)
		assert.NotNil(got)
	})
}

func TestFindInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		ExecutorId: "xxx",
	}
	_, err := FindInstanceByID("i-xxxxx")
	assert.Equal(backend.ErrConnectionNotReady, err)

	withConnect(t, func() {
		got, err := CreateInstance(n)
		assert.NoError(err)
		got2, err := FindInstanceByID(got.Id)
		assert.NoError(err)
		assert.NotNil(got2)
		assert.Equal(got.Id, got2.Id)
		_, err = FindInstanceByID("i-xxxxx")
		assert.Error(err)
	})
}
