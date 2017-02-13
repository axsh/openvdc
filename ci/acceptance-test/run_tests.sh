#!/bin/bash

if [[ "$1" != "RUN" ]]; then
  echo "This script is designed to be run inside of the Docker container provided"\
  "by the Dockerfile in this directory. You most probably don't want to be running"\
  "this manually. Use the 'build_and_run_in_docker.sh' script instead."

  exit 1
fi

set -xe

cat | sudo tee /etc/yum.repos.d/openvdc.repo << EOS
[openvdc]
name=OpenVDC
failovermethod=priority
baseurl=https://ci.openvdc.org/repos/${BRANCH}/${RELEASE_SUFFIX}
enabled=1
gpgcheck=0
EOS

sudo yum install -y openvdc-acceptance-test

/multibox/build.sh

wait_for_port_ready() {
  local ip=$1
  local port=$2
  local started=$(date '+%s')
  while ! (echo "" | nc $ip $port) > /dev/null; do
    echo "Waiting for $ip:$port starts to listen ..."
    sleep 1
    if [[ $(($started + 60)) -le $(date '+%s') ]]; then
      echo "Timed out for ${ip}:${port} becomes ready"
      return 1
    fi
  done
  return 0
}

# gRPC API port
wait_for_port_ready 10.0.100.12 5000

dump_logs() {
  local node=""
  . /multibox/config.source
  for node in ${NODES[@]}
  do
    # Ignore errors to correct logs from all nodes.
    set +e
    echo "journalctl" | SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" /multibox/login.sh $node
  done
}

(
  trap dump_logs EXIT
  # Run the actual tests as axsh user. Root should never be required to run the openvdc command
  su axsh -c "/opt/axsh/openvdc/bin/openvdc-acceptance-test -test.v"
)
