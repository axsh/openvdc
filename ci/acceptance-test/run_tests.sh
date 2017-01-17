#!/bin/bash
set -xe

/multibox/build.sh
/opt/axsh/openvdc/bin/openvdc-acceptance-test -test.v
