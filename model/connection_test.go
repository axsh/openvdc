package model

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	c, err := Connect([]string{os.Getenv("ZK")})
	assert.NoError(err)
	assert.NotNil(c)
	err = Close()
	assert.NoError(err)
}
