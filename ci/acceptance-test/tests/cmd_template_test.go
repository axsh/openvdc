// +build acceptance

package tests

import (
	"testing"
)

//go:generate go-bindata -pkg tests -o fixtures.bindata.go ./fixtures

func init() {
	if err := RestoreAsset("/var/tmp", "fixtures/lxc.json"), 644); err != nil {
		panic(err)
	}
}

func TestCmdTemplateValidate(t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "centos/7/lxc")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "/var/tmp/lxc.json")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
}

func TestCmdTemplateShow(t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "template", "show", "centos/7/lxc")
	RunCmdAndReportFail(t, "openvdc", "template", "show", "/var/tmp/lxc.json")
	RunCmdAndReportFail(t, "openvdc", "template", "show", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
}
