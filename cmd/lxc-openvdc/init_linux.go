package main

import (
	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"log/syslog"
)

func init() {
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_DEBUG, "lxc-openvdc")
	if err != nil {
		logrus.Fatal("Failed to initialize syslog hook: ", err)
	}
	logrus.AddHook(hook)
}
