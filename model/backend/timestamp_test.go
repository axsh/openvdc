package backend

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

//go:generate protoc -I. -I${GOPATH}/src --go_out=. test.proto
func TestTimestampFilter(t *testing.T) {
	assert := assert.New(t)
	tnow := time.Now().UTC()
	f := &TimestampFilter{tnow}

	i := &Timestamp{}
	{
		err := f.OnCreate(i)
		assert.NoError(err)
		assert.NotNil(i.CreatedAt)
		t2, _ := ptypes.Timestamp(i.CreatedAt)
		assert.Equal(tnow, t2)
	}
	{
		err := f.OnUpdate(i)
		assert.NoError(err)
		assert.NotNil(i.UpdatedAt)
		t2, _ := ptypes.Timestamp(i.UpdatedAt)
		assert.Equal(tnow, t2)
	}
}
