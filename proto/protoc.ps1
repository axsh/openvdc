# http://stackoverflow.com/questions/3919798/how-to-check-if-a-cmdlet-exists-in-powershell-at-runtime-via-script
function Check-Command($cmdname)
{
    return [bool](Get-Command -Name $cmdname -ErrorAction SilentlyContinue)
}
$s = Check-Command "protoc"
if ($s -eq $False) {
    Write-Error "Can not find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases"
    Exit 1
}

# Required version of github.com/golang/protobuf
$protoc_go_sha=$(Get-Content .protocgengo.sha)

$skip_goget_protoc = $True
if (-Not $(Check-Command("protoc-gen-go")) ){
    $skip_goget_protoc = $False
} else {
    Push-Location ${env:GOPATH}\src\github.com\golang\protobuf\protoc-gen-go
    if ( -Not $(git rev-list HEAD | Select-String ${protoc_go_sha} -Quiet) ) {
        $skip_goget_protoc = $False
    }
    Pop-Location
}
if ($skip_goget_protoc -eq $False) {
    go get -u -v github.com/golang/protobuf/protoc-gen-go
}

cd (Get-Item $MyInvocation.MyCommand.Path).DirectoryName
# we set option "go_package" so the protoc puts files to the namespace.
protoc -I. -I"${env:GOPATH}/src" --go_out=plugins="grpc:${env:GOPATH}/src" v1.proto
protoc -I. -I"${env:GOPATH}/src" --go_out="${env:GOPATH}/src" model.proto
protoc -I. -I"${env:GOPATH}/src" --go_out=plugins="grpc:${env:GOPATH}/src" executor.proto
Push-Location ..\model\backend
protoc -I. -I"${env:GOPATH}/src" --go_out=. test.proto
Pop-Location
