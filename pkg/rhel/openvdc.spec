# This is a little trick to allow the rpmbuild command to define a suffix for
# development (non stable) versions.
%define release 1
%{?dev_release_suffix:%define release %{dev_release_suffix}}

Name: openvdc
Version: 0.1%{?dev_release_suffix:dev}
Release: %{release}%{?dist}
Summary: Metapackage that depends on all other OpenVDc packages.
Vendor: Axsh Co. LTD <dev@axsh.net>
URL: http://openvdc.org
Source: https://github.com/axsh/openvdc
License: LGPLv3

BuildArch: x86_64

BuildRequires: rpmdevtools

Requires: lxc 
Requires: mesosphere-zookeeper mesos
%{systemd_requires}
Requires: openvdc-cli
Requires: openvdc-executor
Requires: openvdc-scheduler

## This will not work with rpm v. 4.11 (which is what the jenkins vm has!)
##  rpm -v --querytags will give a list of acceptable tags. Suggests: *is*
## an acceptable tag for 4.12, apparently.
#Suggests: mesosphere-zookeeper mesos

%description
This is an empty metapackage that depends on all OpenVNet services and the vnctl client. Just a conventient way to install everything at once on a single machine.

%files
# Metapackage, so no files!


%build
cd "${GOPATH}/src/github.com/axsh/openvdc"
(
  ./build.sh
)

%install
cd "${GOPATH}/src/github.com/axsh/openvdc"
mkdir -p "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
mkdir -p "$RPM_BUILD_ROOT"%{_unitdir}
cp openvdc "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp openvdc-executor "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp openvdc-scheduler "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp pkg/rhel/openvdc-scheduler.service "$RPM_BUILD_ROOT"%{_unitdir}
mkdir -p "$RPM_BUILD_ROOT"/etc/sysconfig
cp pkg/rhel/sysconfig-openvdc "$RPM_BUILD_ROOT"/etc/sysconfig/openvdc


%package cli
Summary: openvdc cli

%description cli
This is an empty message to fulfill the requirement that this file has a "%description" header.

%files cli
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc
%config(noreplace) /etc/sysconfig/openvdc

%package executor
Summary: openvdc executor

%description executor
This is a 'stub'. An appropriate message must be substituted at some point.

%files executor
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-executor

%package scheduler
Summary: openvdc scheduler


%description scheduler
This is a 'stub'. An appropriate message must be substituted at some point.

%files scheduler
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-scheduler
%{_unitdir}/openvdc-scheduler.service


%post
%{systemd_post openvdc-scheduler.service}

%postun
%{systemd_postun openvdc-scheduler.service}

%preun
%{systemd_preun openvdc-scheduler.service}
