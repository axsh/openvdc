
(
    $starting_group "Install golang"
    [[ -f "${TMP_ROOT}/opt/golang/bin/govendor" ]]
    $skip_group_if_unnecessary
    
    (
        $starting_step "Unpack golang archive"
        [[ -f ${TMP_ROOT}/opt/go/bin/go ]]
        $skip_step_if_already_done ; set -ex
        sudo tar -xzf "${TMP_ROOT}/home/go${go_version}.linux-amd64.tar.gz" -C "${TMP_ROOT}/opt/"
    ) ; prev_cmd_failed

    (
        $starting_step "Setup go paths"
        check_env_var "GOROOT" && check_env_var "GOPATH"
        $skip_step_if_already_done ; set -ex
        sudo chroot ${TMP_ROOT} /bin/bash -e <<EOS
             cat <<'EOF' >> /root/.bash_profile
export GOROOT=/opt/go
export GOPATH=/opt/golang
export PATH=\$PATH:\$GOROOT/bin:\$GOPATH/bin
EOF
EOS
    ) ; prev_cmd_failed

    (
        $starting_step "Install go vendor"
        [[ -f "${TMP_ROOT}/opt/golang/bin/govendor" ]]
        $skip_step_if_already_done ; set -ex
        sudo chroot ${TMP_ROOT} /bin/bash -c ". /root/.bash_profile ; go get -u github.com/kardianos/govendor"
    ) ; prev_cmd_failed

) ; prev_cmd_failed
