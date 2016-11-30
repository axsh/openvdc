#!/bin/bash

set -e

VERSION="dev"
SHA=$(git rev-parse --verify HEAD)
BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION=$(go version)
LDFLAGS="-X 'main.version=${VERSION}' -X 'main.sha=${SHA}' -X 'main.builddate=${BUILDDATE}' -X 'main.goversion=${GOVERSION}'"

if [[ ! -x $GOPATH/bin/govendor ]]; then
  go get -u github.com/kardianos/govendor
fi
$GOPATH/bin/govendor sync
if [[ ! -x $GOPATH/bin/go-bindata ]]; then
  go get -u github.com/jteeuwen/go-bindata/...
fi
$GOPATH/bin/go-bindata -pkg registry -o registry/schema.bindata.go schema

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
echo "Done"
