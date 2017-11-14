# This is a little trick to allow the rpmbuild command to define a suffix for
# development (non stable) versions.
%define release 1

Name: openvdc
Version: 0.1%{?dev_release_suffix:dev.%{dev_release_suffix}}
Release: %{release}%{?dist}
Summary: Metapackage that depends on all other OpenVDC packages.
Vendor: Axsh Co. LTD <dev@axsh.net>
URL: http://openvdc.org
Source: https://github.com/axsh/openvdc
License: LGPLv3

BuildArch: x86_64

BuildRequires: rpmdevtools lxc-devel git
# CentOS 7.3 does not have official Go 1.8 package.
#BuildRequires: golang >= 1.8

Requires: mesosphere-zookeeper mesos
Requires: bridge-utils
%{systemd_requires}
Requires: openvdc-cli
Requires: openvdc-executor
Requires: openvdc-executor-lxc
Requires: openvdc-scheduler

## This will not work with rpm v. 4.11 (which is what the jenkins vm has!)
##  rpm -v --querytags will give a list of acceptable tags. Suggests: *is*
## an acceptable tag for 4.12, apparently.
#Suggests: mesosphere-zookeeper mesos

%description
An empty metapackage that depends on all OpenVDC services. Just a conventient way to install everything at once on a single machine.

%files
# Metapackage, so no files!


%build
# rpmbuild resets $PATH so ensure to have "$GOPATH/bin".
export PATH="$PATH:${GOPATH}/bin"
cd "${GOPATH}/src/github.com/axsh/openvdc"
(
  VERSION=%{version} go run ./build.go
)

%install
cd "${GOPATH}/src/github.com/axsh/openvdc"
mkdir -p "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
mkdir -p "$RPM_BUILD_ROOT"%{_unitdir}
mkdir -p "$RPM_BUILD_ROOT"/etc/openvdc
mkdir -p "$RPM_BUILD_ROOT"/etc/openvdc/scripts
mkdir -p "$RPM_BUILD_ROOT"/etc/openvdc/ssh
mkdir -p "$RPM_BUILD_ROOT"/usr/bin
ln -sf /opt/axsh/openvdc/bin/openvdc  "$RPM_BUILD_ROOT"/usr/bin
mkdir -p "$RPM_BUILD_ROOT"/usr/sbin
ln -sf /opt/axsh/openvdc/bin/openvdc-executor  "$RPM_BUILD_ROOT"/usr/sbin
cp openvdc "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp openvdc-executor "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp openvdc-scheduler "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp ci/citest/acceptance-test/tests/openvdc-acceptance-test "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp ci/citest/acceptance-test-esxi/tests/openvdc-acceptance-test-esxi "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp pkg/rhel/openvdc-scheduler.service "$RPM_BUILD_ROOT"%{_unitdir}
cp -r pkg/conf/scripts/ "$RPM_BUILD_ROOT"/etc/openvdc
cp pkg/conf/executor.toml "${RPM_BUILD_ROOT}/etc/openvdc/"
cp pkg/conf/scheduler.toml "${RPM_BUILD_ROOT}/etc/openvdc/"
mkdir -p "$RPM_BUILD_ROOT"/opt/axsh/openvdc/share/mesos-slave
mkdir -p "$RPM_BUILD_ROOT"/opt/axsh/openvdc/share/lxc-templates
cp lxc-openvdc "${RPM_BUILD_ROOT}/opt/axsh/openvdc/share/lxc-templates/lxc-openvdc"
cp qemu-ifup "${RPM_BUILD_ROOT}/opt/axsh/openvdc/share/"
cp qemu-ifdown "${RPM_BUILD_ROOT}/opt/axsh/openvdc/share/"
install -p -t "$RPM_BUILD_ROOT"/opt/axsh/openvdc/share/mesos-slave pkg/conf/mesos-slave/attributes.null pkg/conf/mesos-slave/attributes.lxc pkg/conf/mesos-slave/attributes.qemu pkg/conf/mesos-slave/attributes.esxi

%package cli
Summary: OpenVDC cli

%description cli
The OpenVDC commandline interface.

%files cli
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/usr/bin/openvdc
/opt/axsh/openvdc/bin/openvdc

%package executor
Summary: OpenVDC executor

%description executor
OpenVDC executor common package.

%files executor
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-executor
%dir /opt/axsh/openvdc/share
%dir /opt/axsh/openvdc/share/mesos-slave
/opt/axsh/openvdc/share/mesos-slave/attributes.null
/opt/axsh/openvdc/share/mesos-slave/attributes.lxc
/opt/axsh/openvdc/share/mesos-slave/attributes.qemu
/opt/axsh/openvdc/share/mesos-slave/attributes.esxi
/opt/axsh/openvdc/share/mesos-slave/resources.qemu
/opt/axsh/openvdc/share/mesos-slave/resources.esxi
/opt/axsh/openvdc/share/lxc-templates/lxc-openvdc
/opt/axsh/openvdc/share/qemu-ifup
/opt/axsh/openvdc/share/qemu-ifdown
%dir /etc/openvdc
%dir /etc/openvdc/ssh
/usr/sbin/openvdc-executor

%post executor
test ! -f /etc/openvdc/ssh/host_rsa_key && /usr/bin/ssh-keygen -q -t rsa -f /etc/openvdc/ssh/host_rsa_key -b 4096 -C '' -N '' >&/dev/null;
test ! -f /etc/openvdc/ssh/host_ecdsa_key && /usr/bin/ssh-keygen -q -t ecdsa -f /etc/openvdc/ssh/host_ecdsa_key -C '' -N '' >&/dev/null;
test ! -f /etc/openvdc/ssh/host_ed25519_key && /usr/bin/ssh-keygen -q -t ed25519 -f /etc/openvdc/ssh/host_ed25519_key -C '' -N '' >&/dev/null;

%package executor-null
Summary: OpenVDC executor (null driver)
Requires: openvdc-executor

%post executor-null
if [ -d /etc/mesos-slave ]; then
  if [ ! -f /etc/mesos-slave/attributes ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/attributes.null /etc/mesos-slave/attributes
  fi
fi


%description executor-null
Null driver configuration package for OpenVDC executor.

%files executor-null
%config(noreplace) /etc/openvdc/executor.toml

%package executor-lxc
Summary: OpenVDC executor (LXC driver)
Requires: openvdc-executor
Requires: lxc
# lxc-templates does not resolve its sub dependencies
Requires: lxc-templates wget gpg sed gawk coreutils rsync debootstrap dropbear
Requires: iproute
# Needed for unpacking local images
Requires: tar xz

%description executor-lxc
LXC driver configuration package for OpenVDC executor.

%files executor-lxc
%config(noreplace) /etc/openvdc/executor.toml
%dir /etc/openvdc/scripts
%dir /opt/axsh/openvdc/share/lxc-templates
%config(noreplace) /etc/openvdc/scripts/*

%post executor-lxc
if [ -d /etc/mesos-slave ]; then
  if [ ! -f /etc/mesos-slave/attributes ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/attributes.lxc /etc/mesos-slave/attributes
  fi
fi

cp /opt/axsh/openvdc/share/lxc-templates/lxc-openvdc /usr/share/lxc/templates/lxc-openvdc

%package executor-qemu
Summary: OpenVDC executor (Qemu driver)
Requires: openvdc-executor
Requires: qemu-system-x86
Requires: dosfstools

%description executor-qemu
Qemu driver configuration package for OpenVDC executor

%files executor-qemu
%config(noreplace) /etc/openvdc/executor.toml

%post executor-qemu
if [ -d /etc/mesos-slave ]; then
  if [ ! -f /etc/mesos-slave/attributes ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/attributes.qemu /etc/mesos-slave/attributes
  fi

  if [ ! -f /etc/mesos-slave/resources ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/resources.qemu /etc/mesos-slave/resources
  fi
fi




cp /opt/axsh/openvdc/share/qemu-ifup /etc/
cp /opt/axsh/openvdc/share/qemu-ifdown /etc/

%package executor-esxi
Summary: OpenVDC executor (Esxi driver)
Requires: openvdc-executor
Requires: dosfstools

%description executor-esxi
Esxi driver configuration package for OpenVDC executor

%files executor-esxi
%config(noreplace) /etc/openvdc/executor.toml

%post executor-esxi
if [ -d /etc/mesos-slave ]; then
  if [ ! -f /etc/mesos-slave/attributes ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/attributes.esxi /etc/mesos-slave/attributes
  fi

  if [ ! -f /etc/mesos-slave/resources ]; then
    cp -p /opt/axsh/openvdc/share/mesos-slave/resources.esxi /etc/mesos-slave/resources
  fi
fi

%package scheduler
Summary: OpenVDC scheduler

%description scheduler
This is a 'stub'. An appropriate message must be substituted at some point.

%files scheduler
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-scheduler
%{_unitdir}/openvdc-scheduler.service
%config(noreplace) /etc/openvdc/scheduler.toml

%post
%{systemd_post openvdc-scheduler.service}

%postun
%{systemd_postun openvdc-scheduler.service}

%preun
%{systemd_preun openvdc-scheduler.service}

%package acceptance-test
Summary: The OpenVDC acceptance test used in its CI process.
Requires: openvdc-cli

%description acceptance-test
An acceptance test designed to run on a specifically designed environment. The environment building scripts can be found in the OpenVDC source code repository. The average OpenVDC user will not need to install this.

%files acceptance-test
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-acceptance-test

%package acceptance-test-esxi
Summary: The OpenVDC esxi acceptance test used in its CI process.
Requires: openvdc-cli

%description acceptance-test-esxi
An acceptance test designed to run on a specifically designed environment. The environment building scripts can be found in the OpenVDC source code repository. The average OpenVDC user will not need to install this.

%files acceptance-test-esxi
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc-acceptance-test-esxi
