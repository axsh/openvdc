# openvdc
Extendable Tiny Datacenter Hypervisor on top of Mesos architecture. Wakame-vdc v2 Project.


## Build

Ensure ``$GOPATH`` is set. ``$PATH`` needs to have ``$GOPATH/bin``.

```
go get -u github.com/axsh/openvdc
cd $GOPATH/src/github.com/axsh/openvdc
go run ./internal/build/main.go
```

Build with compile ``proto/*.proto``.

```
go run ./internal/build/main.go -with-gogen
```
