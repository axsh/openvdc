#!/bin/bash

set -xe

VERSION=${VERSION:-"dev"}
SHA=${SHA:-$(git rev-parse --verify HEAD)}
BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')
GOVERSION=$(go version)
BUILDSTAMP="github.com/axsh/openvdc"
LDFLAGS="-X '${BUILDSTAMP}.Version=${VERSION}' -X '${BUILDSTAMP}.Sha=${SHA}' -X '${BUILDSTAMP}.Builddate=${BUILDDATE}' -X '${BUILDSTAMP}.Goversion=${GOVERSION}'"

if [[ -n "$WITH_GOGEN" ]]; then
  if ! type protoc ; then
    echo "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases" >&2
    exit 1
  fi
  if ! type protoc-gen-go; then
    go get -u -v github.com/golang/protobuf/protoc-gen-go
  fi
  if ! type go-bindata; then
    go get -u github.com/jteeuwen/go-bindata/...
  fi
  go generate -v ./api ./model ./registry
fi

if ! type govendor; then
  go get -u github.com/kardianos/govendor
fi
govendor sync

# Determine the default branch reference for registry/github.go
SCHEMA_LAST_COMMIT=${SCHEMA_LAST_COMMIT:-$(git log -n 1 --pretty=format:%H -- schema/ registry/schema.bindata.go)}
if (git rev-list origin/master | grep "${SCHEMA_LAST_COMMIT}") > /dev/null; then
  # Found no changes for resource template/schema on HEAD.
  # so that set preference to the master branch.
  LDFLAGS="${LDFLAGS} -X '${BUILDSTAMP}.GithubDefaultRef=master'"
else
  # Found resource template/schema changes on this HEAD. Switch the default reference branch.
  # Check if $GIT_BRANCH has something once in case of running in Jenkins.
  LDFLAGS="${LDFLAGS} -X '${BUILDSTAMP}.GithubDefaultRef=${GIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}'"
fi

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS -X 'main.DefaultConfPath=/etc/openvdc/executor.toml'" -v ./cmd/openvdc-executor
echo "Done"
