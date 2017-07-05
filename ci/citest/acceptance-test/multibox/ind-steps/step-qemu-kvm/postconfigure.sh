#!/bin/bash

procc_type=
case ${ACCEL_TYPE} in
    vmx) procc_type="intel" ;;
    svm) procc_type="amd" ;;
    *) echo "Invalid hardware acceleration" ; exit 255 ;;
esac

nested="$(cat /sys/module/kvm_${procc_type}/parameters/nested)"
! lsmod | grep -q kvm && {
   echo "Host requires kvm module to be loaded"
   exit 255
}

[[ "${nested}" == "N" || ${nested} -eq 0 ]] && {
   echo "Host requires nested to be turned on" 
   exit 255 
}

(
    $starting_step "Enable kvm module"
    run_ssh root@${IP_ADDR} "lsmod | grep -q kvm"
    $skip_step_if_already_done; set -ex
    run_ssh root@${IP_ADDR} "modprobe kvm_${procc_type}"
) ; prev_cmd_failed
