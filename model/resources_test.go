package model

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"reflect"

	"strings"

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
		assert.Equal(Resource_REGISTERED, got.State)
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
		assert.Equal(got.State, got2.State)
		_, err = Resources(ctx).FindByID("r-xxxxx")
		assert.Error(err)
	})
}

func TestDestroyResource(t *testing.T) {
	assert := assert.New(t)
	err := Resources(context.Background()).Destroy("r-xxxxx")
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		n := &Resource{}
		got, err := Resources(ctx).Create(n)
		assert.NoError(err)
		err = Resources(ctx).Destroy(got.Id)
		assert.NoError(err)
		got2, err := Resources(ctx).FindByID(got.Id)
		assert.NoError(err)
		assert.Equal(Resource_UNREGISTERED, got2.State)
	})
}

var res1 = &Resource{
	Template: &Template{
		TemplateUri: "https://example.com/xxx.json",
		Item: &Template_Lxc{
			Lxc: &LxcTemplate{
				Vcpu:     1,
				MemoryGb: 10,
				LxcImage: &LxcTemplate_Image{
					DownloadUrl: "http://example.com/image.raw",
					Chksum:      "1234567890abcdef",
				},
			}}}}

func TestResource_ResourceTemplate(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/lxc", res1.ResourceTemplate().ResourceName())
}

func ExampleResource_reflection() {
	// Normal type assertion
	rl, ok := res1.Template.Item.(*Template_Lxc)
	fmt.Println(rl.Lxc, ok)
	// Using reflection API
	v := reflect.ValueOf(res1.Template.Item)
	fmt.Println(v.Kind())
	fmt.Println(v.Type().String())
	fmt.Println(
		"ConvertibleTo(*Template_Lxc) ->",
		v.Type().ConvertibleTo(reflect.TypeOf((*Template_Lxc)(nil))),
	)
	fmt.Println(
		"ConvertibleTo(*Template_None) ->",
		v.Type().ConvertibleTo(reflect.TypeOf((*Template_None)(nil))),
	)
	fieldName := strings.TrimPrefix(v.Type().String(), "*model.Template_")
	fmt.Println(fieldName)
	f := v.Elem().FieldByName(fieldName)
	fmt.Println(f.Type().String())
	fmt.Println(f.MethodByName("ResourceName").Call([]reflect.Value{})[0])
	// Output:
	// vcpu:1 memory_gb:10 lxc_image:<download_url:"http://example.com/image.raw" chksum:"1234567890abcdef" >  true
	// ptr
	// *model.Template_Lxc
	// ConvertibleTo(*Template_Lxc) -> true
	// ConvertibleTo(*Template_None) -> false
	// Lxc
	// *model.LxcTemplate
	// vm/lxc
}
