package model

import (
	"testing"
	"time"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func withClusterConnect(t *testing.T, c func(context.Context)) error {
	ze := &backend.ZkEndpoint{}
	if err := ze.Set(unittest.TestZkServer); err != nil {
		t.Fatal("Invalid zookeeper address:", unittest.TestZkServer)
	}
	ctx, err := ClusterConnect(context.Background(), ze)
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
		defer func() {
			err := Cluster(ctx).Unregister("executor1")
			assert.NoError(err)
		}()
		err = Cluster(ctx).Register(n)
		assert.NoError(err, "Should not fail the registration becase here continues session")

		// New cluster connection to run checks from different node.
		withClusterConnect(t, func(ctx context.Context) {
			err := Cluster(ctx).Register(n)
			assert.Error(err, "Should fail the regitration for the same nodeID from different session")
		})
	})
}

func TestCluster_Find(t *testing.T) {
	assert := assert.New(t)
	createdAt, _ := ptypes.TimestampProto(time.Now())
	n := &ExecutorNode{
		Id:       "executor1",
		GrpcAddr: "127.0.0.1:9999",
		Console: &Console{
			Type: Console_SSH,
		},
		LastState: &NodeState{
			State:     NodeState_REGISTERED,
			CreatedAt: createdAt,
		},
		CreatedAt: createdAt,
	}

	var err error
	err = Cluster(context.Background()).Register(n)
	assert.Equal(ErrBackendNotInContext, err)

	withClusterConnect(t, func(ctx context.Context) {
		err := Cluster(ctx).Register(n)
		assert.NoError(err)
		defer func() {
			err := Cluster(ctx).Unregister("executor1")
			assert.NoError(err)
		}()
		n2 := &ExecutorNode{}
		err = Cluster(ctx).Find("executor1", n2)
		assert.NoError(err)
		assert.Equal("executor1", n2.Id)

		/*
			 "proto: bad wiretype for field ..." error is expected here.
				However Protobuf Unmarshaller does not fail with the messages
				have same type fields.

				message A { int32 field1 = 1; }
				message B { int32 myfield = 1; }

				They are recognized as same message because Marshaller generates
				same byte sequence and there is no type info in wireformat.
		*/
		/*
			n3 := &SchedulerNode{}
			err = Cluster(ctx).Find("executor1", n3)
			assert.Error(err, "Should fail to marshall incompatible type")
		*/

		err = Cluster(ctx).Find("unknownXXXX", n2)
		assert.Error(err, "Should fail to find unknown node")
	})
}
