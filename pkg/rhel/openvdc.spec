# This is a little trick to allow the rpmbuild command to define a suffix for
# development (non stable) versions.
%define release 1
%{?dev_release_suffix:%define release %{dev_release_suffix}}

Name: openvdc
Version: 0.9%{?dev_release_suffix:dev}
Release: %{release}%{?dist}
Summary: Metapackage that depends on all other OpenVNet packages.
Vendor: Axsh Co. LTD <dev@axsh.net>
URL: http://openvdc.org
Source: https://github.com/axsh/openvdc
License: LGPLv3

BuildArch: x86_64

BuildRequires: rpmdevtools

%description
This is an empty message to fulfill the requirement that this file has a "%description" header.

%build 
cd "${GOPATH}/src/github.com/axsh/openvdc"
(
  ./build.sh
)

%install
cd "${GOPATH}/src/github.com/axsh/openvdc"
mkdir -p "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin
cp openvdc "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin 
cp openvdc-executor "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin 
cp openvdc-scheduler "$RPM_BUILD_ROOT"/opt/axsh/openvdc/bin 

%files
%dir /opt/axsh/openvdc
%dir /opt/axsh/openvdc/bin
/opt/axsh/openvdc/bin/openvdc
/opt/axsh/openvdc/bin/openvdc-executor
/opt/axsh/openvdc/bin/openvdc-scheduler
