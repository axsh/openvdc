
$zk_host || disable_service "zookeeper"
ssh root@${IP_ADDR} "iptables -F"
