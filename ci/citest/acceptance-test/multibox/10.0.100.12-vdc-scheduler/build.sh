#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))/.." && pwd -P)"
export NODE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP_ROOT="${NODE_DIR}/tmp_root"

. "${ENV_ROOTDIR}/config.source"
. "${NODE_DIR}/vmspec.conf"
. "${ENV_ROOTDIR}/ind-steps/common.source"

scheduler=true

IND_STEPS=(
    "box"
    "ssh"
    "hosts"
    "disable-firewalld"
    "apache"
)

build "${IND_STEPS[@]}"

# This is not part of the ind-steps because we don't want OpenVDC installed in
# the cached images. We want a clean cache without OpenVDC so we can install a
# different version to test every the CI runs.
install_openvdc_yum_repo
install_yum_package_over_ssh "openvdc-scheduler"
enable_service_over_ssh "openvdc-scheduler"

function download_container_image () {
        local img="${1}"
        (
                meta="meta.tar.xz"
                rootfs="rootfs.tar.xz"

                #Image not found, download it.
                if [ ! -f "$IMG_DIR/$imgDirFolder/$meta" ]
                then
                        imgHost="https://uk.images.linuxcontainers.org"
                        imgHostIndex="https://uk.images.linuxcontainers.org/meta/1.0/index-system"
                        imgIndex=$(curl -s "$imgHostIndex")

                        for item in ${imgIndex//\\n/}
                        do
                                if [[ $item =~ $img ]]
                                then
                                        IFS=';' read -a parts <<< "$item"

                                        imageFolder=${parts[-1]}
                                        imgHost+=$imageFolder
                                        imgDirFolder=${img//";"/"-"}

                                        mkdir -p "$IMG_DIR/$imgDirFolder"

                                        printf "\nDownloading file:\n$imgHost$meta\n"
                                        curl -o "$IMG_DIR/$imgDirFolder/$meta" "$imgHost/$meta"

                                        printf "\nDownloading file:\n$imgHost$rootfs\n"
                                        curl -o "$IMG_DIR/$imgDirFolder/$rootfs" "$imgHost/$rootfs"
                                fi
                        done
                fi

                #Copy image to scheduler box.
		scp -i "${ENV_ROOTDIR}/10.0.100.12-vdc-scheduler/sshkey" "$imgHost/$meta" "root@10.0.100.12:/var/www/html/images/"


        ) #; prev_cmd_failed. 
}

for img in "${IMAGES[@]}"
do
        download_container_image "${img}"
done

