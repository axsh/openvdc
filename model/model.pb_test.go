package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImplementsResourceTemplate(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*ResourceTemplate)(nil), new(NoneTemplate))
	assert.Implements((*ResourceTemplate)(nil), new(NullTemplate))
	assert.Implements((*ResourceTemplate)(nil), new(LxcTemplate))
	assert.Implements((*ResourceTemplate)(nil), new(KvmTemplate))
}

func TestImplementsInstanceResource(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*InstanceResource)(nil), new(NullTemplate))
	assert.Implements((*InstanceResource)(nil), new(KvmTemplate))
	assert.Implements((*InstanceResource)(nil), new(LxcTemplate))
}
