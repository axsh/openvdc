package cmd

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type LogFormatter struct {
	DisableTimestamp  bool
	DisableStacktrace bool
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if !f.DisableTimestamp {
		b.WriteString(entry.Time.Format(logrus.DefaultTimestampFormat) + " ")
	}

	fmt.Fprintf(b, "[%s] %s",
		entry.Level.String(),
		entry.Message)
	for _, k := range keys {
		fmt.Fprintf(b, " %s=%s", k, entry.Data[k])
	}
	if !f.DisableStacktrace {
		if err, exists := entry.Data[logrus.ErrorKey]; exists && err != nil {
			type stackTracer interface {
				StackTrace() errors.StackTrace
			}

			if err, ok := err.(stackTracer); ok {
				fmt.Fprintf(b, "\n%v", err)
			}
		}
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}
