#!/bin/bash 


function remove_dir {
    dir_to_rm=$1

    ## "master" and "develop" dris. should NEVER be removed
    for no_rm in "master" "develop"; do
       if [[ "${dir_to_rm}" = "${no_rm}" ]]; then
           echo "Cannot remove \"${no_rm}\". Ignoring."
           return 0
       fi
    done

    echo "rm -rf ${dir_to_rm}  (command not executed)"

}


