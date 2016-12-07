
$scheduler && {
    (
        $starting_step "Starting scheduler on ${vm_name}"
        $skip_step_if_already_done ; set -ex
        ssh root@${IP_ADDR} "/opt/axsh/openvdc/bin/openvdc-scheduler --master=${zk}:5050 --zk=${mesos_master}:2181 --api=${IP_ADDR}:5000"
    ) ; prev_cmd_failed
}
