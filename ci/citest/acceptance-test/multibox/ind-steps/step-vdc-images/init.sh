#!/bin/bash

for img in "${IMAGES[@]}" ; do
    item=$(grep "$img" <<< "$imgIndex")
    IFS=';' read -a parts <<< "$item"
    IFS=';' eval 'folders=($item)'
    imgRemoteDir="${parts[-1]}"

    prepare_container_image "${img}" "${folders[0]}/${folders[1]}/${folders[2]}" "$imgRemoteDir"
    add_container_image "${img}" "${folders[0]}/${folders[1]}/${folders[2]}"
done

(
    # resync again after images have been downloaded
    $starting_step "Synching guestroot for ${vm_name}"
    # This step is set to false by default and we rely on rsyncs -u flag
    # to take care of keeping the files updated
    false
    $skip_step_if_already_done; set -ex
    sudo rsync -rv "${NODE_DIR}/guestroot/" "${TMP_ROOT}"
) ; prev_cmd_failed
