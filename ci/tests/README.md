
Running acceptance/CI tests under ``ci/tests``:

```
cd ci/tests
go test -tags=acceptance ./...
```

Writing acceptance/CI tests:

 - Each test file written in Go. And the file name ends with ``_test.go``.
 - Place ``// +build acceptance`` magic comment at the first line.

Minimal code example:

```go
// +build acceptance

package tests

import "testing"

func TestScenario1(t *testing.T) {
	t.Log("TestScenario1")
}
```
