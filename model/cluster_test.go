package model

import (
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model/backend"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func withClusterConnect(t *testing.T, c func(context.Context)) error {

	ctx, err := ClusterConnect(context.Background(), []string{unittest.TestZkServer})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := ClusterClose(ctx)
		if err != nil {
			t.Error("Failed ClusterClose:", err)
		}
	}()
	err = InstallSchemas(GetClusterBackendCtx(ctx).(backend.ModelSchema))
	if err != nil {
		t.Fatal(err)
	}
	c(ctx)
	return err
}

func TestClusterNode(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*ClusterNode)(nil), &ExecutorNode{})
	assert.Implements((*ClusterNode)(nil), &SchedulerNode{})
}

func TestCluster_Register(t *testing.T) {
	assert := assert.New(t)
	n := &ExecutorNode{
		Id: "executor1",
	}

	var err error
	err = Cluster(context.Background()).Register(n)
	assert.Equal(ErrBackendNotInContext, err)

	withClusterConnect(t, func(ctx context.Context) {
		err := Cluster(ctx).Register(n)
		assert.NoError(err)
		err = Cluster(ctx).Register(n)
		assert.Error(err, "Should fail the registration for duplicated node")
	})
}

func TestCluster_Find(t *testing.T) {
	assert := assert.New(t)
	n := &ExecutorNode{
		Id: "executor1",
	}

	var err error
	err = Cluster(context.Background()).Register(n)
	assert.Equal(ErrBackendNotInContext, err)

	withClusterConnect(t, func(ctx context.Context) {
		err := Cluster(ctx).Register(n)
		assert.NoError(err)
		n2 := &ExecutorNode{}
		err = Cluster(ctx).Find("executor1", n2)
		assert.NoError(err)
		assert.Equal("executor1", n2.Id)

		n3 := &SchedulerNode{}
		err = Cluster(ctx).Find("executor1", n3)
		assert.Error(err, "Should fail to marshall incompatible type")

		err = Cluster(ctx).Find("unknownXXXX", n2)
		assert.Error(err, "Should fail to find unknown node")
	})
}
