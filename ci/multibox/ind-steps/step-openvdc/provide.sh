
$scheduler && {
    (
        $starting_step "Starting scheduler on ${vm_name}"
	ssh root@${IP_ADDR} "systemctl status openvdc-scheduler | grep -q running"
        $skip_step_if_already_done ; set -ex
        ssh root@${IP_ADDR} "systemctl start openvdc-scheduler"
    ) ; prev_cmd_failed
}
