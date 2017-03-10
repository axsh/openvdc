// +build acceptance

package tests

import "testing"

func TestCmdTemplateValidate(t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "centos/7/lxc")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "./templates/centos/7/lxc.json")
	RunCmdAndReportFail(t, "openvdc", "template", "validate", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
}

func TestCmdTemplateShow(t *testing.T) {
	RunCmdAndReportFail(t, "openvdc", "template", "show", "centos/7/lxc")
	RunCmdAndReportFail(t, "openvdc", "template", "show", "./templates/centos/7/lxc.json")
	RunCmdAndReportFail(t, "openvdc", "template", "show", "https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json")
}
