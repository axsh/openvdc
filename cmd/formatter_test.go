package cmd

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

func newLogger(w io.Writer) *logrus.Logger {
	log := logrus.New()
	log.Formatter = new(LogFormatter)
	log.Out = w
	return log
}

func TestLogFomatter(t *testing.T) {
	buf := new(bytes.Buffer)
	log := newLogger(buf)

	log.Info("test")
	ok, err := regexp.Match(
		"^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} \\[INFO\\] github.com/axsh/openvdc/cmd/formatter_test.go:\\d+ test",
		buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Does not match: ", buf.String())
	}
}

func TestLogFomatter_StackTrace(t *testing.T) {
	buf := new(bytes.Buffer)
	log := newLogger(buf)

	err := errors.New("err1")
	err = errors.WithStack(err)

	log.WithError(err).Error("test")

	lines := strings.SplitN(buf.String(), "\n", 2)
	ok, err := regexp.MatchString(
		"^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} \\[ERROR\\] github.com/axsh/openvdc/cmd/formatter_test.go:\\d+ test error=err1",
		lines[0])
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Does not match log line: ", lines[0])
	}
	ok, err = regexp.MatchString(
		"^\\s*github\\.com/axsh/openvdc/cmd\\.TestLogFomatter_StackTrace",
		lines[1])
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Does not match stack trace line: ", lines[1])
	}
}
