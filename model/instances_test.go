package model

import (
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func withConnect(t *testing.T, c func(context.Context)) error {
	ctx, err := Connect(context.Background(), []string{os.Getenv("ZK")})
	if err != nil {
		t.Fatal(err)
	}
	defer Close(ctx)
	c(ctx)
	return err
}

func TestCreateInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		ExecutorId: "xxx",
	}

	_, err := Instances(context.Background()).Create(n)
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		got, err := Instances(ctx).Create(n)
		assert.NoError(err)
		assert.NotNil(got)
	})
}

func TestFindInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		ExecutorId: "xxx",
	}
	_, err := Instances(context.Background()).FindByID("i-xxxxx")
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		got, err := Instances(ctx).Create(n)
		assert.NoError(err)
		got2, err := Instances(ctx).FindByID(got.Id)
		assert.NoError(err)
		assert.NotNil(got2)
		assert.Equal(got.Id, got2.Id)
		_, err = Instances(ctx).FindByID("i-xxxxx")
		assert.Error(err)
	})
}
