
$VERSION="dev"
$SHA=$(git rev-parse --verify HEAD)
$BUILDDATE=Get-Date -Format "yyyy/MM/dd HH:mm:ss zzz"
$GOVERSION=$(go version)
$LDFLAGS="-X 'main.version=${VERSION}' -X 'main.sha=${SHA}' -X 'main.builddate=${BUILDDATE}' -X 'main.goversion=${GOVERSION}'"

Invoke-Expression "$(Join-Path ${env:GOPATH}\bin govendor.exe -Resolve) sync"

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
Write-Host "Done"
