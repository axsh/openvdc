# openvdc
Extendable Tiny Datacenter Hypervisor on top of Mesos architecture. Wakame-vdc v2 Project.


## Build

Ensure to have ``$GOPATH`` and ``$GOPATH/bin`` is included in ``$PATH``.

```
go get -u github.com/axsh/openvdc
go get -u github.com/kardianos/govendor
cd $GOPATH/src/github.com/axsh/openvdc
govendor sync
./build.sh
```
