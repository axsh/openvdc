#!/bin/bash

: "${image_type:?"should be set"}"

case "${image_type}" in
    qemu) # stores directly in cache until we can use compressed files
        IMAGES=( "${QEMU_IMAGES[@]}" )
        LOCAL_STORAGE_PATH="${NODE_DIR}/guestroot/var/cache/qemu"
        ;;
    lxc) # stores on local image server
        IMAGES=( "${LXC_IMAGES[@]}" )
        LOCAL_STORAGE_PATH="${NODE_DIR}/guestroot/var/www/html/images"
        ;;
    *)
        echo "Unknown image type: ${image_type}."
		exit 255
        ;;
esac

for img in "${IMAGES[@]}" ; do
   cache_${image_type}_image  "${img}"
done

