
$VERSION="dev"
$SHA=$(git rev-parse --verify HEAD)
$BUILDDATE=Get-Date -Format "yyyy/MM/dd HH:mm:ss zzz"
$GOVERSION=$(go version)
$LDFLAGS="-X 'main.version=${VERSION}' -X 'main.sha=${SHA}' -X 'main.builddate=${BUILDDATE}' -X 'main.goversion=${GOVERSION}'"

Invoke-Expression "$(Join-Path ${env:GOPATH}\bin govendor.exe -Resolve) sync"
$modtime=$(git log -n 1 --date=raw --pretty=format:%cd -- schema/).split(" ", 1)
Invoke-Expression "$(Join-Path ${env:GOPATH}\bin go-bindata.exe -Resolve) -modtime ${modtime} -pkg registry -o registry\schema.bindata.go schema"

# Determine the default branch reference for registry/github.go
$SCHEMA_LAST_COMMIT=$(git log -n 1 --pretty=format:%H -- schema/ registry/schema.bindata.go)
git rev-list origin/master | Select-String "${SCHEMA_LAST_COMMIT}"
if ( $? -eq 0 ) {
  # Found no changes for resource template/schema on HEAD.
  # so that set preference to the master branch.
  $LDFLAGS="${LDFLAGS} -X 'registry.GithubDefaultRef=master'"
} else {
  # Found resource template/schema changes on this HEAD. Switch the default reference branch.
  # Check if $GIT_BRANCH has something once in case of running in Jenkins.
  $branch = if( $env:GIT_BRANCH ){
      $env:GIT_BRANCH
  }else{
      $(git rev-parse --abbrev-ref HEAD)
  }
  $LDFLAGS="${LDFLAGS} -X 'registry.GithubDefaultRef=${branch}'"
}

go build -ldflags "$LDFLAGS" -v ./cmd/openvdc
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-scheduler
go build -ldflags "$LDFLAGS" -v ./cmd/openvdc-executor
Write-Host "Done"
