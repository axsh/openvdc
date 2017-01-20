package model

import (
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model/backend"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)

	ze := &backend.ZkEndpoint{}
	if err := ze.Set(unittest.TestZkServer); err != nil {
		t.Fatal("Invalid zookeeper address:", unittest.TestZkServer)
	}
	ctx, err := Connect(context.Background(), ze)
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
