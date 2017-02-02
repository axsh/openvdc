#!/bin/bash

set -e

. ../prune_branches.sh

time_limit=${TIME_LIMIT:-14}  ## Days. Set this to give the "deadline". 
                              ## All branches older than this a removed.

function docker_image_date {
    image=${1?"No image passed to docker_image_date!"}

    creation_date=$(docker inspect --format='{{.Created}}' --type=image ${image})

    ## The date is in the format: yyyy-mm-ddThh:mm:ss.xxxx  
    ## We want: yyyymmdd
    creation_date=${creation_date%T*}

    echo ${creation_date//-/}    ## Remove the '-' between yyyy & mm, mm & dd
}

#-------------------------------------------------------------------------#
# main()

cutoff_date=$(get_cutoff_date ${time_limit})   ## Images older than this are removed

## Remove all directories whose branch (on git) no longer exists
## or which has not beenm pushed to within $time_limit days.
for docker_image in $(docker images -q | sort -u); do
   image_date=$(docker_image_date ${docker_image})

   if [[ "${image_date}" < "${cutoff_date}" ]]; then
       echo "docker rmi \"${docker_image}\""
       docker rmi "${docker_image}"
   fi

done
 
exit 0   ## Explicit notice: We are done.
