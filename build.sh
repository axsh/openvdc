#!/bin/bash

set -e

VERSION="dev"
SHA=$(git rev-parse --verify HEAD)
BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION=$(go version)
LDFLAGS="-X 'main.version=${VERSION}' -X 'main.sha=${SHA}' -X 'main.builddate=${BUILDDATE}' -X 'main.goversion=${GOVERSION}'"
# During development, assume that the executor binary locates in the build directory.
EXECUTOR_PATH=$(pwd)/openvdc-executor

go get -u github.com/kardianos/govendor
govendor sync

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -ldflags "-X 'github.com/axsh/openvdc/scheduler.ExecutorPath=${EXECUTOR_PATH}'" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
echo "Done"
