FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
RUN yum install -y epel-release

RUN yum install -y git go 
ENV GOPATH=/var/tmp/go 
ENV PATH=$PATH:$GOPATH/bin
RUN mkdir $GOPATH

RUN mkdir -p $GOPATH/src/github.com/axsh/openvdc
RUN go get -u github.com/kardianos/govendor

RUN yum install -y http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm
RUN yum install -y mesosphere-zookeeper
RUN yum install -y lxc lxc-devel
