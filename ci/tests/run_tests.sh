#!/bin/bash

set -xe

if [[ ! -x $GOPATH/bin/govendor ]]; then
  go get -u github.com/kardianos/govendor
fi
$GOPATH/bin/govendor sync -v

go test -tags=acceptance $(go list ./... | grep -v /vendor/)
