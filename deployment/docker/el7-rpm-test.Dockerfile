FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
RUN yum install -y epel-release
RUN yum install -y http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm
RUN yum install -y iproute nc
ADD deployment/docker/yum.repo/dev.repo /etc/yum.repos.d/
RUN mkdir -p /etc/zookeeper/conf; echo "SERVER_JVMFLAGS=\"-Djava.net.preferIPv4Stack=True\"" > /etc/zookeeper/conf/java.env
