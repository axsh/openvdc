#!/bin/bash

set -e

if ! type protoc ; then
  echo "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases"
  exit 1
fi

# Required version of github.com/golang/protobuf
protoc_go_sha="8ee79997227bf9b34611aee7946ae64735e6fd93"

skip_goget_protoc=1
if ! type protoc-gen-go; then
  skip_goget_protoc=0
elif ! (cd $GOPATH/src/github.com/golang/protobuf/protoc-gen-go; git rev-list HEAD | grep "${protoc_go_sha}") > /dev/null; then
  skip_goget_protoc=0
fi
if [[ $skip_goget_protoc -eq 0 ]]; then
  go get -u -v github.com/golang/protobuf/protoc-gen-go
fi

protoc -I . --go_out=plugins=grpc:../api v1.proto
protoc -I . --go_out=../model model.proto
