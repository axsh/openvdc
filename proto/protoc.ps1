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
$s = Check-Command("protoc-gen-go")
if ($s -eq $False) {
    go get -u -v github.com/golang/protobuf/protoc-gen-go
}

cd (Get-Item $MyInvocation.MyCommand.Path).DirectoryName
protoc -I . --go_out=plugins=grpc:. v1.proto
