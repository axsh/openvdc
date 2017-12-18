package main

import (
	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"

	_ "github.com/axsh/openvdc/hypervisor/esxi"
	_ "github.com/axsh/openvdc/hypervisor/lxc"
	_ "github.com/axsh/openvdc/hypervisor/null"
	_ "github.com/axsh/openvdc/hypervisor/qemu"
)

func init() {
	// forward log messages to local syslog.
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_DEBUG, "vdc-executor")
	if err != nil {
		logrus.Fatal("Failed to initialize syslog hook: ", err)
	}
	logrus.AddHook(hook)
}
