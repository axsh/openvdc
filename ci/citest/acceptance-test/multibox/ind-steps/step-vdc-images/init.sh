#!/bin/bash

for img in "${IMAGES[@]}" ; do
    item=$(grep "$img" <<< "$imgIndex")
    IFS=';' read -a parts <<< "$item"
    IFS=';' eval 'folders=($item)'
    imgRemoteDir="${parts[-1]}"

    prepare_container_image "${img}" "${folders[0]}/${folders[1]}/${folders[2]}" "$imgRemoteDir"
    add_container_image "${img}" "${folders[0]}/${folders[1]}/${folders[2]}"
done

