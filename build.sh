#!/bin/bash

set -xe

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

modtime=$(git log -n 1 --date=raw --pretty=format:%cd -- schema/ | cut -d' ' -f1)
$GOPATH/bin/go-bindata -modtime "${modtime}" -pkg registry -o registry/schema.bindata.go schema

# Determine the default branch reference for registry/github.go
SCHEMA_LAST_COMMIT=${SCHEMA_LAST_COMMIT:-$(git log -n 1 --pretty=format:%H -- schema/ registry/schema.bindata.go)}
if (git rev-list origin/master | grep "${SCHEMA_LAST_COMMIT}") > /dev/null; then
  # Found no changes for resource template/schema on HEAD.
  # so that set preference to the master branch.
  LDFLAGS="${LDFLAGS} -X 'registry.GithubDefaultRef=master'"
else
  # Found resource template/schema changes on this HEAD. Switch the default reference branch.
  # Check if $GIT_BRANCH has something once in case of running in Jenkins.
  LDFLAGS="${LDFLAGS} -X 'registry.GithubDefaultRef=${GIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}'"
fi

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
echo "Done"
