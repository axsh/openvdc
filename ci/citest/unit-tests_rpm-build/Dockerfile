FROM centos:7

RUN ["yum", "install", "-y", "epel-release"]
RUN ["yum", "install", "-y",  "git", "go", "http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm", "mesosphere-zookeeper", "lxc", "lxc-devel"]

RUN groupadd -r axsh && useradd -d /home/axsh -g axsh axsh && mkdir /home/axsh/go
USER axsh
WORKDIR /home/axsh

ENV GOPATH=/home/axsh/go
ENV PATH=:$PATH:$GOPATH:/bin

RUN go get -u github.com/kardianos/govendor

VOLUME $GOPATH/src/github.com/axsh/openvdc

LABEL "jp.axsh.vendor"="Axsh Co. LTD"  \
      "jp.axsh.project"="OpenVDC" \
      "jp.axsh.task"="Unit tests and RPM build" \
      "jp.axsh.branch"="$BRANCH" \
      "jp.axsh.release_suffix"="$RELEASE_SUFFIX" \
      "jp.axsh.buildtime"="$ISO8601_TIMESTAMP" \
      "jp.axsh.build_url"="$BUILD_URL" \
      "jp.axsh.git_commit"="$LONG_SHA"

ENTRYPOINT ["go/src/github.com/axsh/openvdc/ci/citest/unit-tests_rpm-build/run_tests_and_rpmbuild.sh"]
