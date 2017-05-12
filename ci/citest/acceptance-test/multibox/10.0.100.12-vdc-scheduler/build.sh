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
		printf "\nPreparing to download image: $img \n"
                meta="meta.tar.xz"
                rootfs="rootfs.tar.xz"

                imgHost="https://uk.images.linuxcontainers.org"
                imgHostIndex="https://uk.images.linuxcontainers.org/meta/1.0/index-system"
                imgIndex=$(curl -s "$imgHostIndex")

                for item in ${imgIndex//\\n/}
                do
                	if [[ $item =~ $img ]]
               	 	then
				IFS=';' read -a parts <<< "$item"

				imgRemoteDir=${parts[-1]}
				imgHost+=$imgRemoteDir

				IFS=';' eval 'folders=($item)'
				imgSpec="${folders[0]}/${folders[1]}/${folders[2]}"

				if [ ! -d "$IMG_DIR/$imgSpec" ]; then
					mkdir -p "$IMG_DIR/$imgSpec"
				else
					printf "\nDirectory already exists: $IMG_DIR/$imgSpec\n"
				fi

				if [ ! -f "$IMG_DIR/$imgSpec/$meta" ]; then
					printf "\nDownloading file:\n$imgHost$meta\n"
					curl -o "$IMG_DIR/$imgSpec/$meta" "$imgHost/$meta"
				else
					printf "\nFile already exists, skipping download: $IMG_DIR/$imgSpec/$meta\n"
				fi

				if [ ! -f "$IMG_DIR/$imgSpec/$rootfs" ]; then
					printf "\nDownloading file:\n$imgHost$rootfs\n"
					curl -o "$IMG_DIR/$imgSpec/$rootfs" "$imgHost/$rootfs"
				else
					printf "\nFile already exists, skipping download: $IMG_DIR/$imgSpec/$rootfs\n"
				fi
                	fi
               	done

		printf "\nCreating folder ${IP_ADDR}:/var/www/html/images/$imgSpec\n"
		ssh -o StrictHostKeyChecking=no -i "${ENV_ROOTDIR}/10.0.100.12-vdc-scheduler/sshkey" "root@${IP_ADDR}" "mkdir -p /var/www/html/images/$imgSpec"
	
		if [ -f "$IMG_DIR/$imgSpec/$meta" ]; then
			printf "\nCopying file: $IMG_DIR/$imgSpec/$meta\n"
			scp -o StrictHostKeyChecking=no -i "${ENV_ROOTDIR}/10.0.100.12-vdc-scheduler/sshkey" "$IMG_DIR/$imgSpec/$meta" "root@${IP_ADDR}:/var/www/html/images/$imgSpec"
		else
			printf "\nFile not found: $IMG_DIR/$imgSpec/$meta\n"
		fi

		if [ -f "$IMG_DIR/$imgSpec/$rootfs" ]; then
                        printf "\nCopying file: $IMG_DIR/$imgSpec/$rootfs\n"
                        scp -o StrictHostKeyChecking=no -i "${ENV_ROOTDIR}/10.0.100.12-vdc-scheduler/sshkey" "$IMG_DIR/$imgSpec/$rootfs" "root@${IP_ADDR}:/var/www/html/images/$imgSpec"
                else
                        printf "\nFile not found: $IMG_DIR/$imgSpec/$rootfs\n"
                fi
        )
}

for img in "${IMAGES[@]}"
do
        download_container_image "${img}"
done

