FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
RUN yum install -y yum-utils
# epel-release.rpm from CentOS/extra contains deprecated index for mirror sites.
RUN yum install -y http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-10.noarch.rpm

RUN yum install -y git
RUN curl -L https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz | tar -C /usr/local -xzf -
RUN yum install -y gcc
ENV GOPATH=/var/tmp/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
RUN mkdir $GOPATH

RUN mkdir -p $GOPATH/src/github.com/axsh/openvdc
RUN go get -u github.com/kardianos/govendor

RUN yum install -y http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm
RUN yum install -y mesosphere-zookeeper
RUN yum install -y lxc lxc-devel
