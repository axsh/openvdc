FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
RUN yum install -y yum-utils createrepo rpm-build rpmdevtools rsync sudo
#RUN yum install -y make gcc gcc-c++ git \
#   mariadb-devel sqlite-devel libpcap-devel
RUN yum install -y make git go
RUN go get -u github.com/kardianos/govendor
RUN mkdir /var/tmp/openvdc
#ADD deployment/docker/yum.repo/dev.repo /etc/yum.repos.d/
# Only enables "openvdc-third-party" repo.
#RUN yum-config-manager --disable openvdc
