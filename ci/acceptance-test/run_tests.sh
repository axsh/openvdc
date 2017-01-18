#!/bin/bash
#TODO: Require a commandline option and output a usage is not supplied
#      Just stopping people from running this locally without thinking. :p
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
