#!/bin/bash

set -e

if ! type protoc ; then
  echo "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases"
  exit 1
fi
if ! type protoc-gen-go; then
  go get -u -v github.com/golang/protobuf/protoc-gen-go
fi
protoc -I . --go_out=plugins=grpc:. v1.proto
protoc -I . --go_out=. model.proto
