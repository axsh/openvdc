package model

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

func withConnect(t *testing.T, c func(context.Context)) error {
	ze := &backend.ZkEndpoint{}
	if err := ze.Set(unittest.TestZkServer); err != nil {
		t.Fatal("Invalid zookeeper address:", unittest.TestZkServer)
	}
	ctx, err := Connect(context.Background(), ze)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := Close(ctx)
		if err != nil {
			t.Error("Failed to Close:", err)
		}
	}()
	err = InstallSchemas(GetBackendCtx(ctx).(backend.ModelSchema))
	if err != nil {
		t.Fatal(err)
	}
	c(ctx)
	return err
}

func TestCreateInstance(t *testing.T) {
	assert := assert.New(t)
	n := &Instance{
		SlaveId: "xxx",
	}

	var err error
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
		SlaveId: "xxx",
		Template: &Template{
			Item: &Template_Lxc{
				Lxc: lxcTmpl1,
			},
		}}

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
	err := Instances(context.Background()).UpdateState("i-xxxxx", InstanceState_REGISTERED)
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		n := &Instance{
			SlaveId: "xxx",
			Template: &Template{
				Item: &Template_Lxc{
					Lxc: lxcTmpl1,
				},
			}}
		got, err := Instances(ctx).Create(n)
		assert.NoError(err)
		assert.Equal(InstanceState_REGISTERED, got.GetLastState().State)
		assert.Error(Instances(ctx).UpdateState(got.GetId(), InstanceState_SHUTTINGDOWN))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_QUEUED))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_STARTING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_RUNNING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_STOPPING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_STOPPED))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_STARTING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_RUNNING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_REBOOTING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_RUNNING))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_SHUTTINGDOWN))
		assert.NoError(Instances(ctx).UpdateState(got.GetId(), InstanceState_TERMINATED))
	})
}

func TestUpdateInstance(t *testing.T) {
	assert := assert.New(t)
	err := Instances(context.Background()).Update(&Instance{Id: "i-xxxx"})
	assert.Equal(ErrBackendNotInContext, err)

	withConnect(t, func(ctx context.Context) {
		n := &Instance{
			SlaveId: "xxx",
			Template: &Template{
				Item: &Template_Lxc{
					Lxc: lxcTmpl1,
				},
			}}
		err = Instances(ctx).Update(n)
		assert.Error(err)
		assert.Equal(ErrInvalidID, err, "Empty ID object should get ErrInvalidID")

		got, err := Instances(ctx).Create(n)
		assert.NoError(err)
		got.Template.GetLxc().Vcpu = 100
		err = Instances(ctx).Update(got)
		got2, err := Instances(ctx).FindByID(got.Id)
		assert.NoError(err)
		assert.Equal(int32(100), got2.Template.GetLxc().Vcpu)
	})
}

func TestInstanceState_ValidateNextState(t *testing.T) {
	assert := assert.New(t)

	s := &InstanceState{
		State: InstanceState_REGISTERED,
	}
	assert.NoError(s.ValidateNextState(InstanceState_QUEUED))
	s.State = InstanceState_QUEUED
	assert.NoError(s.ValidateNextState(InstanceState_STARTING))
	s.State = InstanceState_STARTING
	assert.NoError(s.ValidateNextState(InstanceState_RUNNING))
	s.State = InstanceState_RUNNING
	assert.NoError(s.ValidateNextState(InstanceState_SHUTTINGDOWN))
	assert.NoError(s.ValidateNextState(InstanceState_STOPPING))
	s.State = InstanceState_STOPPING
	assert.NoError(s.ValidateNextState(InstanceState_STOPPED))
	s.State = InstanceState_SHUTTINGDOWN
	assert.NoError(s.ValidateNextState(InstanceState_TERMINATED))
	s.State = InstanceState_TERMINATED
	assert.Error(s.ValidateNextState(InstanceState_TERMINATED))
	assert.Error(s.ValidateNextState(InstanceState_RUNNING))
}

func TestInstanceState_ValidateGoalState(t *testing.T) {
	assert := assert.New(t)

	s := &InstanceState{
		State: InstanceState_REGISTERED,
	}
	assert.NoError(s.ValidateGoalState(InstanceState_QUEUED))
	s.State = InstanceState_QUEUED
	assert.NoError(s.ValidateGoalState(InstanceState_RUNNING))
	assert.NoError(s.ValidateGoalState(InstanceState_STOPPED))
	s.State = InstanceState_STARTING
	assert.NoError(s.ValidateGoalState(InstanceState_RUNNING))
	s.State = InstanceState_RUNNING
	assert.NoError(s.ValidateGoalState(InstanceState_TERMINATED))
	assert.NoError(s.ValidateGoalState(InstanceState_STOPPED))
	s.State = InstanceState_STOPPING
	assert.NoError(s.ValidateGoalState(InstanceState_STOPPED))
	s.State = InstanceState_SHUTTINGDOWN
	assert.NoError(s.ValidateGoalState(InstanceState_TERMINATED))
	s.State = InstanceState_TERMINATED
	assert.Error(s.ValidateGoalState(InstanceState_TERMINATED))
	assert.Error(s.ValidateGoalState(InstanceState_RUNNING))
}

var lxcTmpl1 = &LxcTemplate{
	Vcpu:     1,
	MemoryGb: 10,
	LxcImage: &LxcTemplate_Image{
		DownloadUrl: "http://example.com/image.raw",
		Chksum:      "1234567890abcdef",
	},
}

var res1 = &Instance{
	CreatedAt: new(timestamp.Timestamp),
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

func TestInstance_ResourceTemplate(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/lxc", res1.ResourceTemplate().ResourceName())
}

func ExampleResourceTemplate_reflection() {
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
		"ConvertibleTo(*Template_Null) ->",
		v.Type().ConvertibleTo(reflect.TypeOf((*Template_Null)(nil))),
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
	// ConvertibleTo(*Template_Null) -> false
	// Lxc
	// *model.LxcTemplate
	// vm/lxc
}
