package cmd

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestLogFomatter(t *testing.T) {
	buf := new(bytes.Buffer)
	log := logrus.New()
	log.Formatter = new(LogFormatter)
	log.Out = buf

	log.Info("test")
	ok, err := regexp.Match(
		"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\+\\d{2}:\\d{2} \\[INFO\\] github.com/axsh/openvdc/cmd/formatter_test.go:\\d+ test",
		buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Does not match: ", buf.String())
	}
}
