#!/bin/bash

set -e

if ! type protoc ; then
  echo "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases"
  exit 1
fi

# Required version of github.com/golang/protobuf
protoc_go_sha=$(cat .protocgengo.sha)

skip_goget_protoc=1
if ! type protoc-gen-go; then
  skip_goget_protoc=0
elif ! (cd $GOPATH/src/github.com/golang/protobuf/protoc-gen-go; git rev-list HEAD | grep "${protoc_go_sha}") > /dev/null; then
  skip_goget_protoc=0
fi
if [[ $skip_goget_protoc -eq 0 ]]; then
  go get -u -v github.com/golang/protobuf/protoc-gen-go
fi

# we set option "go_package" so the protoc puts files to the namespace.
protoc -I. -I"${GOPATH}/src" --go_out=plugins=grpc:$GOPATH/src v1.proto

