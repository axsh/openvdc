package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		ExecutorId: "xxx",
	}
	got, err := CreateInstance(n)
	assert.NoError(err)
	assert.NotNil(got)
	println(got)
}
