package model

import (
	"os"
	"testing"

	"golang.org/x/net/context"
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

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	ctx, err := Connect(context.Background(), []string{testZkServer})
	assert.NoError(err)
	assert.NotNil(ctx)
	err = Close(ctx)
	assert.NoError(err)
}
