#!/bin/bash   -x

set -e

VERSION="dev"
SHA=$(git rev-parse --verify HEAD)
BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION=$(go version)
LDFLAGS="-X 'main.version=${VERSION}' -X 'main.sha=${SHA}' -X 'main.builddate=${BUILDDATE}' -X 'main.goversion=${GOVERSION}'"
# During development, assume that the executor binary locates in the build directory.
EXECUTOR_PATH=$(pwd)/openvdc-executor

#export GOPATH=$PWD
$GOPATH/bin/govendor sync

if [[ $(ls -1 ./vendor | wc -l) -eq 1 ]]; then
  echo "ERROR: ./vendor has not been setup yet." >&2
  exit 1
fi

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -ldflags "-X 'github.com/axsh/openvdc/scheduler.ExecutorPath=${EXECUTOR_PATH}'" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
echo "Done"
