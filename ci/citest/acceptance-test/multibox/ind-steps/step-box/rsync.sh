#!/bin/bash

(
    # Doing this step as a seprate stage so we can perform tasks for each step
    # while making sure that it happens before we start running yum.
    # If we install software using yum, their configuration timestampts will
    # be newer than the ones in guestroot, causing those files to not be
    # transfered because of the -u flag, 
    $starting_step "Synching guestroot for ${vm_name}"
    # This step is set to false by default and we rely on rsyncs -u flag
    # to take care of keeping the files updated
    false
    $skip_step_if_already_done; set -ex
    sudo rsync -rv "${NODE_DIR}/guestroot/" "${TMP_ROOT}"
) ; prev_cmd_failed
