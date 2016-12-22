package model

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestFilterCreatedAt(t *testing.T) {
	assert := assert.New(t)
	tnow, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return
	}

	i := &Instance{}
	FilterCreatedAt(i, tnow)
	assert.NotNil(i.CreatedAt)
}
