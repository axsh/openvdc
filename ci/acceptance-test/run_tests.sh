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

# Run the actual tests as axsh user. Root should never be required to run the openvdc command
su axsh -c "/opt/axsh/openvdc/bin/openvdc-acceptance-test -test.v"
