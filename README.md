# openvdc
Extendable Tiny Datacenter Hypervisor on top of Mesos architecture. Wakame-vdc v2 Project.


## Build

Ensure ``$GOPATH`` is set. ``$PATH`` needs to have ``$GOPATH/bin``.

The developer package of lxc is also required. (lxc-devel from epel-release on redhat systems, or lxc-dev on debian systems) package is also required.

```
go get -u github.com/axsh/openvdc
cd $GOPATH/src/github.com/axsh/openvdc
go run ./build.go
```

Build with compile ``proto/*.proto``.

```
go run ./build.go -with-gogen
```
