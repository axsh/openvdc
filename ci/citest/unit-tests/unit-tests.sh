#!/bin/bash

set -ex -o pipefail
SCRIPT_DIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"

CID=
docker_rm() {
  if [[ -z "${CID}" ]]; then
    return 0
  fi
  if [[ -n "$LEAVE_CONTAINER" ]]; then
     if [[ "${LEAVE_CONTAINER}" != "0" ]]; then
        echo "Skip to clean container: ${CID}"
        return 0
     fi
  fi
  docker rm -f "${CID}"
}
trap 'docker_rm' EXIT

BUILD_ENV_PATH=${1:?"ERROR: env file is not given."}
if [[ -n "${BUILD_ENV_PATH}" && ! -f "${BUILD_ENV_PATH}" ]]; then
  echo "ERROR: Can't find the file: ${BUILD_ENV_PATH}" >&2
  exit 1
fi

set -a
. ${BUILD_ENV_PATH}
set +a

if [[ -n "$JENKINS_HOME" ]]; then
  # openvdc-axsh/branch1/el7
  img_tag=$(echo "unit-tests.${JOB_NAME}/${BUILD_OS}" | tr '/' '.')
else
  img_tag="unit-tests.openvdc.$(git rev-parse --abbrev-ref HEAD).${BUILD_OS}"
fi
# Docker 1.10 fails with uppercase image tag name. need letter case translation.
# https://github.com/docker/docker/issues/20056
img_tag="${img_tag,,}"

docker build -t "${img_tag}" -f "${SCRIPT_DIR}/${BUILD_OS}-unit-tests.Dockerfile" .
CID=$(docker run --add-host="devrepo:${IPV4_DEVREPO:-192.168.56.60}" -d ${BUILD_ENV_PATH:+--env-file $BUILD_ENV_PATH} "${img_tag}")


docker cp . "${CID}:/var/tmp/go/src/github.com/axsh/openvdc"

## Run unit tests
docker exec $CID /bin/bash -c "/usr/bin/env"
docker exec $CID /bin/bash -c "cd /var/tmp/go/src/github.com/axsh/openvdc;  govendor sync"
docker exec $CID /bin/bash -c 'cd /var/tmp/go/src/github.com/axsh/openvdc;  ZK=127.0.0.1 GITHUB_DEFAULT_REF=${GIT_BRANCH} go test $(go list ./... | grep -v /vendor/)'
docker exec $CID /bin/bash -c 'cd /var/tmp/go/src/github.com/axsh/openvdc;  go test -v -tags=terraform $(go list ./terraform/...)'
