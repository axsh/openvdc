FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
# epel-release.rpm from CentOS/extra contains deprecated index for mirror sites.
RUN yum install -y http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-8.noarch.rpm
RUN yum install -y yum-utils go git createrepo


ENV GOPATH=/var/tmp/go PATH=$PATH:$GOPATH/bin
RUN mkdir $GOPATH
RUN go get -u github.com/kardianos/govendor
RUN mkdir -p /var/tmp/go/src/github.com/axsh/openvdc
RUN mkdir -p /var/tmp/rpmbuild/SOURCES
RUN ln -s ${GOPATH}/src/github.com/axsh/openvdc  /var/tmp/rpmbuild/SOURCES/openvdc
ENV WORK_DIR=/var/tmp/rpmbuild
