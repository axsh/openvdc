#!/bin/bash
# Run build process in docker container.

set -ex -o pipefail

OPENVDC_ROOT="$(cd "$(dirname $(readlink -f "$0"))/../../" && pwd -P)"
SCRIPT_DIR="${OPENVDC_ROOT}/deployment/docker"

CID=
TMPDIR=$(mktemp -d)
function docker_rm() {
    if [[ -z "$CID" ]]; then
        return 0
    fi
    if [[ -n "$LEAVE_CONTAINER" ]]; then
        if [[ "${LEAVE_CONTAINER}" != "0" ]]; then
            echo "Skip to clean container: ${CID}"
            return 0
        fi
    fi
    docker rm -f "$CID"
}

trap "docker_rm; rm -rf ${TMPDIR}" EXIT

BUILD_ENV_PATH=${1:?"ERROR: env file is not given."}

if [[ -n "${BUILD_ENV_PATH}" && ! -f "${BUILD_ENV_PATH}" ]]; then
  echo "ERROR: Can't find the file: ${BUILD_ENV_PATH}" >&2
  exit 1
fi

echo "COMMIT_ID=$(git rev-parse HEAD)" >> ${BUILD_ENV_PATH}
# /tmp is memory file system on docker.
echo "WORK_DIR=/var/tmp/rpmbuild" >> ${BUILD_ENV_PATH}

# http://stackoverflow.com/questions/19331497/set-environment-variables-from-file
set -a
. ${BUILD_ENV_PATH}
set +a

function docker_cp() {
  local cid=${2%:*}
  if [[ -z $cid ]]; then
    # container -> host
    docker cp $1 $2
  else
    # host -> container. Docker 1.7 or earlier does not support.
    docker cp $1 $2 || {
      local path=${2#*:}
      tar -cO $1 | docker exec -i "${cid}" bash -c "tar -xf - -C ${path}"
    }
  fi
}

if [[ -n "$JENKINS_HOME" ]]; then
  # openvnet-axsh/branch1/el7
  img_tag=$(echo "${JOB_NAME}/${BUILD_OS}" | tr '/' '.')
  # $BUILD_CACHE_DIR/openvnet-axsh/el7/0123abcdef.tar.gz
  build_cache_base="${BUILD_CACHE_DIR}/${BUILD_OS}/${JOB_NAME%/*}"
else
  img_tag="openvdc.$(git rev-parse --abbrev-ref HEAD).${BUILD_OS}"
  build_cache_base="${BUILD_CACHE_DIR}"
fi

### This is the location on the dh machine where the openvdc yum repo is to be placed
RPM_ABSOLUTE=/var/www/html/openvdc-repos/

docker build -t "${img_tag}" -f "${SCRIPT_DIR}/${BUILD_OS}.Dockerfile" .
CID=$(docker run --add-host="devrepo:${IPV4_DEVREPO:-192.168.56.60}" ${BUILD_ENV_PATH:+--env-file $BUILD_ENV_PATH} -d "${img_tag}")

# Upload checked out tree to the container.
docker_cp "${OPENVDC_ROOT}/." "${CID}:/var/tmp/go/src/github.com/axsh/openvdc"

# Upload build cache if found.
if [[ -n "$BUILD_CACHE_DIR" && -d "${build_cache_base}" ]]; then
  for f in $(ls ${build_cache_base}); do
    cached_commit=$(basename $f)
    cached_commit="${cached_commit%.*}"
    if git rev-list "${COMMIT_ID}" | grep "${cached_commit}" > /dev/null; then
      echo "FOUND build cache ref ID: ${cached_commit}"
      cat "${build_cache_base}/$f" | docker cp - "${CID}:/"
      break;
    fi
  done
fi
# Run build script

# We'll need to wrap this into a condtional at some point: Are we building dev version or stable version?
#docker exec -t "${CID}" /bin/bash -c "cd /var/tmp/go/src/github.com/axsh/openvdc ; rpmbuild -ba --define \"_topdir ${WORK_DIR}\" pkg/rhel/openvdc.spec"

docker exec -t "${CID}" /bin/bash -c "cd /var/tmp/go/src/github.com/axsh/openvdc ; rpmbuild -ba --define \"_topdir ${WORK_DIR}\"  --define \"dev_release_suffix ${RELEASE_SUFFIX}\"  pkg/rhel/openvdc.spec"

# Build the yum repository
docker exec -t "${CID}" /bin/bash -c "mkdir -p /var/tmp/${RELEASE_SUFFIX}/" 
docker exec  -t "${CID}" /bin/bash -c "mv /var/tmp/rpmbuild/RPMS/*  /var/tmp/${RELEASE_SUFFIX}/" 

docker exec -t "${CID}" /bin/bash -c "cd /var/tmp/${RELEASE_SUFFIX}/ ; createrepo . "

if [[ -n "$BUILD_CACHE_DIR" ]]; then
    if [[ ! -d "$BUILD_CACHE_DIR" || ! -w "$BUILD_CACHE_DIR" ]]; then
        echo "ERROR: BUILD_CACHE_DIR '${BUILD_CACHE_DIR}' does not exist or not writable." >&2
        exit 1
    fi
    if [[ ! -d "${build_cache_base}" ]]; then
      mkdir -p "${build_cache_base}"
    fi
    docker cp "${SCRIPT_DIR}/build-cache.list" "${CID}:/var/tmp/build-cache.list"
    docker exec "${CID}" tar cO --directory=/ --files-from=/var/tmp/build-cache.list > "${build_cache_base}/${COMMIT_ID}.tar"
    # Clear build cache files which no longer referenced from Git ref names (branch, tags)
    git show-ref --head --dereference | awk '{print $1}' > "${TMPDIR}/sha.a"
    for i in $(git reflog show | head -10 | awk '{print $2}'); do
      git rev-parse "$i"
    done >> "${TMPDIR}/sha.a"
    (cd "${build_cache_base}"; ls *.tar) | cut -d '.' -f1 > "${TMPDIR}/sha.b"
    # Set operation: B - A
    join -v 2 <(sort -u ${TMPDIR}/sha.a) <(sort -u ${TMPDIR}/sha.b) | while read i; do
      echo "Removing build cache: ${build_cache_base}/${i}.tar"
      rm -f "${build_cache_base}/${i}.tar" || :
    done
fi
# Pull compiled yum repository
# $SSH_REMOTE is set within the Jenkins configuration ("Manage Jenkins" --> "Configure System")
docker cp "${CID}:/var/tmp/${RELEASE_SUFFIX}/" - | $SSH_REMOTE tar xf - -C "${RPM_ABSOLUTE}"
