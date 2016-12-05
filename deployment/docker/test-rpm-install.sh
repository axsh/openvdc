#!/bin/bash

set -ex -o pipefail

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
  img_tag=$(echo "rpm-install.${JOB_NAME}/${BUILD_OS}" | tr '/' '.')
else
  img_tag="rpm-install.openvdc.$(git rev-parse --abbrev-ref HEAD).${BUILD_OS}"
fi

docker build -t "${img_tag}" -f "./deployment/docker/${BUILD_OS}-rpm-test.Dockerfile" .
CID=$(docker run --add-host="devrepo:${IPV4_DEVREPO:-192.168.56.60}" -d ${BUILD_ENV_PATH:+--env-file $BUILD_ENV_PATH} "${img_tag}")
docker exec -t $CID /bin/sh -c "echo '${RELEASE_SUFFIX}' > /etc/yum/vars/ovn_release_suffix"
docker exec $CID yum install -y openvdc

## Setup
docker exec $CID systemctl enable zookeeper
docker exec $CID systemctl enable mesos-master
docker exec $CID systemctl enable mesos-slave
docker exec $CID systemctl enable openvdc-scheduler

docker exec $CID systemctl start zookeeper
started=$(date '+%s')
while ! (echo "" | docker exec $CID nc 127.0.0.1 2181) > /dev/null; do
  echo "Waiting for zookeeper starts to listen '0.0.0.0:2181' ..."
  sleep 1
  if [[ $(($started + 60)) -le $(date '+%s') ]]; then
    echo "Timed out for zookeeper becomes ready."
    exit 1
  fi
done
docker exec $CID systemctl start mesos-master
#docker exec $CID systemctl start mesos-slave
docker exec $CID systemctl start openvdc-scheduler

docker exec $CID systemctl status zookeeper
docker exec $CID systemctl status mesos-master
#docker exec $CID systemctl status mesos-slave
docker exec $CID systemctl status openvdc-scheduler

