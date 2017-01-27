[CmdletBinding()]
param(
    [parameter(Mandatory=$False,Position=1)]
    [string]$VERSION = "dev",
    [parameter(Mandatory=$False)]
    [switch]$WITH_GOGEN
)

$ErrorActionPreference = "Stop"

$SHA=$(git rev-parse --verify HEAD)
$BUILDDATE=Get-Date -Format "yyyy/MM/dd HH:mm:ss zzz"
$GOVERSION=$(go version)
$BUILDSTAMP="github.com/axsh/openvdc"
$LDFLAGS="-X '${BUILDSTAMP}.Version=${VERSION}' -X '${BUILDSTAMP}.Sha=${SHA}' -X '${BUILDSTAMP}.Builddate=${BUILDDATE}' -X '${BUILDSTAMP}.Goversion=${GOVERSION}'"

if( $WITH_GOGEN ) {
    try{
        Get-Command -Name protoc | Out-Null
    } catch [System.Management.Automation.CommandNotFoundException] {
        Write-Error "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases"
        Exit 1
    }
    try{
        Get-Command -Name protoc-gen-go | Out-Null
    } catch [System.Management.Automation.CommandNotFoundException] {
       go get -u -v github.com/golang/protobuf/protoc-gen-go
    }
    try{
        Get-Command -Name go-bindata | Out-Null
    } catch [System.Management.Automation.CommandNotFoundException] {
        go get -u github.com/jteeuwen/go-bindata/...
    }
    go generate -v ./api ./model ./registry
}

try{
    Get-Command -Name govendor | Out-Null
} catch [System.Management.Automation.CommandNotFoundException] {
    go get -u github.com/kardianos/govendor
}
govendor sync

$branch="master"
if( $env:APPVEYOR_REPO_BRANCH ){
    $branch = $env:APPVEYOR_REPO_BRANCH
}else{
    # Determine the default branch reference for registry/github.go
    $SCHEMA_LAST_COMMIT=$(git log -n 1 --pretty=format:%H -- schema/ registry/schema.bindata.go)
    $f=$(git rev-list origin/master | Select-String "${SCHEMA_LAST_COMMIT}" -ErrorAction "Continue")
    $branch = if ( $f -ne $null ) {
        # Found no changes for resource template/schema on HEAD.
        # so that set preference to the master branch.
        "master"
    } else {
        # Found resource template/schema changes on this HEAD. Switch the default reference branch.
        # Check if $BRANCH_NAME has something once in case of running in Jenkins.
        if( $env:BRANCH_NAME ){
            $env:BRANCH_NAME
        }else{
            $(git rev-parse --abbrev-ref HEAD)
        }
    }
}
$LDFLAGS="${LDFLAGS} -X '${BUILDSTAMP}.GithubDefaultRef=${branch}'"

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS -X 'main.DefaultConfPath=C:\\openvdc\\etc\\scheduler.conf'" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
Write-Host "Done"
