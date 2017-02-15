package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const DefaultTimestampFormat = "2006-01-02 15:04:05"

type LogFormatter struct {
	DisableTimestamp  bool
	DisableStacktrace bool
	DisableFileLine   bool
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if !f.DisableTimestamp {
		b.WriteString(entry.Time.Format(DefaultTimestampFormat) + " ")
	}

	fmt.Fprintf(b, "[%s]", strings.ToUpper(entry.Level.String()))
	if !f.DisableFileLine {
		file, line := firstCaller()
		fmt.Fprintf(b, " %s:%d", file, line)
	}
	fmt.Fprintf(b, " %s", entry.Message)

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, " %s=%v", k, entry.Data[k])
	}
	if !f.DisableStacktrace {
		if err, exists := entry.Data[logrus.ErrorKey]; exists && err != nil {
			type stackTracer interface {
				StackTrace() errors.StackTrace
			}

			if err, ok := err.(stackTracer); ok {
				fmt.Fprintf(b, "\n%+v", err.StackTrace())
			}
		}
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}

const pkgLogrus = "github.com/Sirupsen/logrus"

func firstCaller() (string, int) {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		file = trimGOPATH(fn.Name(), file)
		// Prefix match does not work for vendored pkgLogrus
		if strings.Index(file, pkgLogrus) == -1 {
			return file, line
		}
	}
	return "", 0
}

// github.com/pkg/errors/stack.go
func trimGOPATH(name, file string) string {
	// Here we want to get the source file path relative to the compile time
	// GOPATH. As of Go 1.6.x there is no direct way to know the compiled
	// GOPATH at runtime, but we can infer the number of path segments in the
	// GOPATH. We note that fn.Name() returns the function name qualified by
	// the import path, which does not include the GOPATH. Thus we can trim
	// segments from the beginning of the file path until the number of path
	// separators remaining is one more than the number of path separators in
	// the function name. For example, given:
	//
	//    GOPATH     /home/user
	//    file       /home/user/src/pkg/sub/file.go
	//    fn.Name()  pkg/sub.Type.Method
	//
	// We want to produce:
	//
	//    pkg/sub/file.go
	//
	// From this we can easily see that fn.Name() has one less path separator
	// than our desired output. We count separators from the end of the file
	// path until it finds two more than in the function name and then move
	// one character forward to preserve the initial path segment without a
	// leading separator.
	const sep = "/"
	goal := strings.Count(name, sep) + 2
	i := len(file)
	for n := 0; n < goal; n++ {
		i = strings.LastIndex(file[:i], sep)
		if i == -1 {
			// not enough separators found, set i so that the slice expression
			// below leaves file unmodified
			i = -len(sep)
			break
		}
	}
	// get back to 0 or trim the leading separator
	file = file[i+len(sep):]
	return file
}
