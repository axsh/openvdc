FROM centos:7

#TODO: Set user
#TODO: Set workdir to user home
#TODO: Set name

ENV BRANCH=acceptance-test RELEASE_SUFFIX=current REBUILD=false

VOLUME /data2

RUN ["yum", "install", "-y", "epel-release"]
RUN ["yum", "install", "-y", "rsync", "bridge-utils", "qemu-kvm", "qemu-system-x86", "parted", "sudo", "openssh-clients", "nmap-ncat"]

ADD ["openvdc.repo", "/etc/yum.repos.d/openvdc.repo"]
RUN ["yum", "install", "-y", "openvdc-acceptance-test"]

ADD ["multibox", "/multibox"]
ADD ["run_tests.sh", "run_tests.sh"]

ENTRYPOINT ["./run_tests.sh"]
