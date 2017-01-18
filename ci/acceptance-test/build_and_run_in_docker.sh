#!/bin/bash

set -ex -o pipefail

whereami="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"

BUILD_ENV_PATH=${1:?"ERROR: env file is not given."}
if [[ -n "${BUILD_ENV_PATH}" && ! -f "${BUILD_ENV_PATH}" ]]; then
  echo "ERROR: Can't find the file: ${BUILD_ENV_PATH}" >&2
  exit 1
fi

set -a
. ${BUILD_ENV_PATH}
set +a

DATA_DIR="${DATA_DIR:-/data2}"

repo_and_tag="openvdc/acceptance-test:${BRANCH}.${RELEASE_SUFFIX}"

sudo docker build -t "${repo_and_tag}" --build-arg BRANCH="${BRANCH}" \
                                  --build-arg RELEASE_SUFFIX="${RELEASE_SUFFIX}" \
                                  --build-arg REBUILD="${REBUILD}" \
                                  "${whereami}"

sudo docker run --privileged -v "${DATA_DIR}":/data "${repo_and_tag}"
