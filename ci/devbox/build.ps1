$ErrorActionPreference = "Stop"

if(!(Get-Command -Name .\packer.exe )){
    Write-Error "packer command not found. Please install from https://packer.io/"
    exit 1
}

if(!(Get-Command -Name 7z.exe )){
    Write-Error "7z command not found. Please install from http://www.7-zip.org/"
    exit 1
}

$box_url="https://atlas.hashicorp.com/bento/boxes/centos-7.3/versions/2.3.2/providers/virtualbox.box"
$box_tmp="boxtemp\7.3"

if(!(Test-Path -Path $box_tmp)){
    New-Item -ItemType directory -Path $box_tmp | Out-Null
}

if(!(Test-Path -Path "${box_tmp}\t.box")){
    Write-Host "Downloading ${box_url} to ${box_tmp}\t.box"
    Invoke-WebRequest -Uri $box_url -OutFile "${box_tmp}\t.box"
}

pushd $box_tmp
7z -y e "t.box"
7z -y x "t"
popd

$env:HOST_SWITCH="VirtualBox Host-Only Ethernet Adapter"
.\packer build -force devbox-centos7.json
