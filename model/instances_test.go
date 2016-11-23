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

	var err error
	_, err = Instances(context.Background()).Create(n)
	assert.Equal(ErrInstanceMissingResource, err)
	n.ResourceId = "r-xxxx"
	_, err = Instances(context.Background()).Create(n)
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
		ResourceId: "r-xxxx",
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

func TestUpdateStateInstance(t *testing.T) {
	assert := assert.New(t)
	err := Instances(context.Background()).UpdateState("i-xxxxx", InstanceState_INSTANCE_REGISTERED)
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		n := &Instance{
			ExecutorId: "xxx",
			ResourceId: "r-xxxx",
		}
		got, err := Instances(ctx).Create(n)
		assert.NoError(err)
		assert.Equal(InstanceState_INSTANCE_REGISTERED, got.GetState())
		assert.Error(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_TERMINATED))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_QUEUED))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_STARTING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_RUNNING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_STOPPING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_STOPPED))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_STARTING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_RUNNING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_SHUTTINGDOWN))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_INSTANCE_TERMINATED))
	})
}
