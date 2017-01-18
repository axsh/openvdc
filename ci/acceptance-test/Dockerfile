FROM centos:7

MAINTAINER "Axsh Co. LTD"

#TODO: Set name
#TODO: Set labels
#TODO: Install openvdc-cli in the container... openvdc-acceptance-test should depend on it

VOLUME /data2

RUN ["yum", "install", "-y", "epel-release"]
RUN ["yum", "install", "-y", "rsync", "bridge-utils", "qemu-kvm", "qemu-system-x86", "parted", "sudo", "openssh-clients", "nmap-ncat"]

RUN groupadd -r axsh && useradd -r -g axsh axsh
WORKDIR /home/axsh

ARG BRANCH
ARG RELEASE_SUFFIX
ARG REBUILD

# Set the ARGs to ENV because otherwise they're not visible to the run_tests.sh script
ENV BRANCH=${BRANCH:-master} RELEASE_SUFFIX=${RELEASE_SUFFIX:-current} REBUILD=${REBUILD:-false}

COPY ["multibox", "/multibox"]
COPY ["run_tests.sh", "run_tests.sh"]

ENTRYPOINT ["./run_tests.sh"]
