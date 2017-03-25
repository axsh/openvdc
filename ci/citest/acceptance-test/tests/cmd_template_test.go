// +build acceptance

package tests

import (
	"testing"

	"github.com/tidwall/gjson"
)

//go:generate go-bindata -pkg tests -o fixtures.bindata.go ./fixtures

func init() {
	if err := RestoreAsset("/var/tmp", "fixtures/lxc.json"); err != nil {
		panic(err)
	}
}

func TestCmdTemplateValidate(t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "centos/7/lxc")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "/var/tmp/fixtures/lxc.json")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
}

func TestCmdTemplateShow(t *testing.T) {
	{
		stdout, _ := RunCmdAndReportFail(t, "openvdc", "template", "show", "centos/7/lxc")
		js := gjson.ParseBytes(stdout.Bytes())
		if !js.Get("lxcTemplate").Exists() {
			t.Error("lxcTemplate key not found")
		}
	}

	{
		stdout, _ := RunCmdAndReportFail(t, "openvdc", "template", "show", "/var/tmp/fixtures/lxc.json")
		js := gjson.ParseBytes(stdout.Bytes())
		if !js.Get("lxcTemplate").Exists() {
			t.Error("lxcTemplate key not found")
		}
	}

	{
		stdout, _ := RunCmdAndReportFail(t, "openvdc", "template", "show", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
		js := gjson.ParseBytes(stdout.Bytes())
		if !js.Get("lxcTemplate").Exists() {
			t.Error("lxcTemplate key not found")
		}
	}
}
