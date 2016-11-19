package model

import (
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	ctx, err := Connect(context.Background(), []string{os.Getenv("ZK")})
	assert.NoError(err)
	assert.NotNil(ctx)
	err = Close(ctx)
	assert.NoError(err)
}
