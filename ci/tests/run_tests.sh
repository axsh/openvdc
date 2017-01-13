#!/bin/bash

set -xe

# Some temporary hacks to get openvdc in the GOPATH on the CI.
# Will fix this once we start running the acceptance tests in docker containers.
OPENVDC_ROOT="$(cd "$(dirname $(readlink -f "$0"))/../../" && pwd -P)"
OPENVDC_LINKNAME="${GOPATH}/src/github.com/axsh/$(basename ${OPENVDC_ROOT})"

function cleanup_symlink {
  rm -f "${OPENVDC_LINKNAME}"
}
trap cleanup_symlink EXIT

if [ "${PWD##$GOPATH}" == "${PWD}"  ]; then
  mkdir -p "${GOPATH}/src/github.com/axsh/"
  ln -s "${OPENVDC_ROOT}" "${OPENVDC_LINKNAME}"
fi
cd "${OPENVDC_LINKNAME}/ci/tests"
# End of temporary hacks

if [[ ! -x $GOPATH/bin/govendor ]]; then
  go get -u github.com/kardianos/govendor
fi
$GOPATH/bin/govendor sync -v

go test -tags=acceptance $(go list ./... | grep -v /vendor/)
