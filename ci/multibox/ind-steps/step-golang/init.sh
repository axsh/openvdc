
(
    $starting_step "Download golang"
    [[ -f "${TMP_ROOT}/home/go${go_version}.linux-amd64.tar.gz" ]]
    $skip_step_if_already_done ; set -ex
    sudo wget "https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz" \
         -P "${TMP_ROOT}/home"
) ;prev_cmd_failed

