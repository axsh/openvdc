FROM centos:7

#TODO: Set user
#TODO: Set workdir to user home
#TODO: Set name
#TODO: Install openvdc-cli in the container... openvdc-acceptance-test should depend on it

VOLUME /data2

RUN ["yum", "install", "-y", "epel-release"]
RUN ["yum", "install", "-y", "rsync", "bridge-utils", "qemu-kvm", "qemu-system-x86", "parted", "sudo", "openssh-clients", "nmap-ncat"]

ARG BRANCH
ARG RELEASE_SUFFIX
ARG REBUILD

# Set the ARGs to ENV because otherwise they're not visible to the run_tests.sh script
ENV BRANCH=${BRANCH:-master} RELEASE_SUFFIX=${RELEASE_SUFFIX:-current} REBUILD=${REBUILD:-false}

ADD ["multibox", "/multibox"]
ADD ["run_tests.sh", "run_tests.sh"]

ENTRYPOINT ["./run_tests.sh"]
