#!/bin/bash

set -e

if ! type packer; then
  echo "packer command not found. Please install from https://packer.io/" >&2
  exit 1
fi
#box_url="${1:?ERROR: Require to set download .box URL}"
box_url="https://atlas.hashicorp.com/bento/boxes/centos-7.3/versions/2.3.2/providers/virtualbox.box"
box_tmp="${2:-boxtemp/7.3}"

# ignore duplicating dir
mkdir -p $box_tmp  || :

(
  cd $box_tmp
  if [ -f './.etag' ]; then
      etag=$(cat ./.etag)
  fi
  curl --dump-header box.header ${etag:+-H "If-None-Match: ${etag}"} -L -o "t.box" "${box_url}"
  cat box.header | awk 'BEGIN {FS=": "}/^ETag/{print $2}' > .etag
  rm -f box.header
  tar -xzf t.box
)

export HOST_SWITCH=${HOST_SWITCH:-vboxnet0}
packer build -force devbox-centos7.json
