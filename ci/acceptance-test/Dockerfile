FROM centos:7

VOLUME /data2

RUN ["yum", "install", "-y", "epel-release"]
RUN ["yum", "install", "-y", "rsync", "bridge-utils", "qemu-kvm", "qemu-system-x86", "parted", "sudo", "openssh-clients", "nmap-ncat"]

ENV BRANCH=acceptance-test
ENV RELEASE_SUFFIX=current
ENV REBUILD=false

ADD ["openvdc.repo", "/etc/yum.repos.d/openvdc.repo"]
RUN ["yum", "install", "-y", "openvdc-acceptance-test"]

ADD ["multibox", "/multibox"]
ADD ["run_tests.sh", "run_tests.sh"]

ENTRYPOINT ["./run_tests.sh"]
