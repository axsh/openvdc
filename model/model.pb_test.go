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
	assert.Implements((*ResourceTemplate)(nil), new(QemuTemplate))
	assert.Implements((*ResourceTemplate)(nil), new(EsxiTemplate))
}

func TestImplementsInstanceResource(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*InstanceResource)(nil), new(NullTemplate))
	assert.Implements((*InstanceResource)(nil), new(QemuTemplate))
	assert.Implements((*InstanceResource)(nil), new(LxcTemplate))
	assert.Implements((*InstanceResource)(nil), new(EsxiTemplate))
}

func TestImplementsConsoleAuthAttributes(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*ConsoleAuthAttributes)(nil), new(NullTemplate))
	assert.Implements((*ConsoleAuthAttributes)(nil), new(QemuTemplate))
	assert.Implements((*ConsoleAuthAttributes)(nil), new(LxcTemplate))
}
