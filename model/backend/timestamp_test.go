package backend

import (
	"testing"
	"time"

	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestTimestampFilter(t *testing.T) {
	assert := assert.New(t)
	tnow := time.Now()
	f := &TimestampFilter{tnow}

	i := &model.Instance{}
	err := f.OnCreate(i)
	assert.NoError(err)
	assert.NotNil(i.CreatedAt)
	t2, _ := ptypes.Timestamp(i.CreatedAt)
	assert.Equal(tnow, t2)
}
