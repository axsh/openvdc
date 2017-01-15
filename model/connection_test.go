package model

import (
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	ctx, err := Connect(context.Background(), []string{unittest.TestZkServer})
	assert.NoError(err)
	assert.NotNil(ctx)
	err = Close(ctx)
	assert.NoError(err)
}

func TestClusterConnect(t *testing.T) {
	assert := assert.New(t)
	ctx, err := ClusterConnect(context.Background(), []string{unittest.TestZkServer})
	assert.NoError(err)
	assert.NotNil(ctx)
	err = ClusterClose(ctx)
	assert.NoError(err)
}
