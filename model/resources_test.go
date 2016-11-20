package model

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestCreateResource(t *testing.T) {
	assert := assert.New(t)
	n := &Resource{}

	_, err := Resources(context.Background()).Create(n)
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		got, err := Resources(ctx).Create(n)
		assert.NoError(err)
		assert.NotNil(got)
	})
}

func TestFindResource(t *testing.T) {
	assert := assert.New(t)
	n := &Resource{}
	_, err := Resources(context.Background()).FindByID("r-xxxxx")
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		got, err := Resources(ctx).Create(n)
		assert.NoError(err)
		got2, err := Resources(ctx).FindByID(got.Id)
		assert.NoError(err)
		assert.NotNil(got2)
		assert.Equal(got.Id, got2.Id)
		_, err = Resources(ctx).FindByID("r-xxxxx")
		assert.Error(err)
	})
}
