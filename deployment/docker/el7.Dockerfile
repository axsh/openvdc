FROM centos:7
WORKDIR /var/tmp
ENTRYPOINT ["/sbin/init"]
RUN yum install -y yum-utils go git epel-release createrepo


ENV GOPATH=/var/tmp/go PATH=$PATH:$GOPATH/bin
RUN mkdir $GOPATH
RUN go get -u github.com/kardianos/govendor
RUN mkdir -p /var/tmp/go/src/github.com/axsh/openvdc
RUN mkdir -p /var/tmp/rpmbuild/SOURCES
RUN ln -s ${GOPATH}/src/github.com/axsh/openvdc  /var/tmp/rpmbuild/SOURCES/openvdc
ENV WORK_DIR=/var/tmp/rpmbuild
