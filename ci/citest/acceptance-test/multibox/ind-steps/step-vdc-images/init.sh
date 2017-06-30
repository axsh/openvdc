#!/bin/bash

case "${image_type}" in
    qemu) # stores directly in cache until we can use compressed files
        IMAGES=( "${QEMU_IMAGES[@]}" )
        LOCAL_STORAGE_PATH="${NODE_DIR}/guestroot/var/cache/qemu"
        ;;
    lxc) # stores on local image server
        IMAGES=( "${LXC_IMAGES[@]}" )
        LOCAL_STORAGE_PATH="${NODE_DIR}/guestroot/var/ww/html/images"
        ;;
    *)
        echo "Unknown image type."
        ;;
esac

for img in "${IMAGES[@]}" ; do
   cache_${image_type}_image  "${img}"
done

