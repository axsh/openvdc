[CmdletBinding()]
param(
    [parameter(Mandatory=$False,Position=1)]
    [string]$VERSION = "dev"
)

$ErrorActionPreference = "Stop"

$SHA=$(git rev-parse --verify HEAD)
$BUILDDATE=Get-Date -Format "yyyy/MM/dd HH:mm:ss zzz"
$GOVERSION=$(go version)
$BUILDSTAMP="github.com/axsh/openvdc"
$LDFLAGS="-X '${BUILDSTAMP}.Version=${VERSION}' -X '${BUILDSTAMP}.Sha=${SHA}' -X '${BUILDSTAMP}.Builddate=${BUILDDATE}' -X '${BUILDSTAMP}.Goversion=${GOVERSION}'"

Invoke-Expression "$(Join-Path ${env:GOPATH}\bin govendor.exe -Resolve) sync"
$modtime=$(git log -n 1 --date=raw --pretty=format:%cd -- schema/).split(" ")[0]
Invoke-Expression "$(Join-Path ${env:GOPATH}\bin go-bindata.exe -Resolve) -mode 420 -modtime ${modtime} -pkg registry -o registry\schema.bindata.go schema"

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
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
Write-Host "Done"
