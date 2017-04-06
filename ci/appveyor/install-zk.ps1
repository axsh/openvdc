$ErrorActionPreference = "Stop"

Set-Location -Path $Args[0]

Invoke-WebRequest -Uri http://www-us.apache.org/dist/zookeeper/zookeeper-3.4.9/zookeeper-3.4.9.tar.gz -OutFile zookeeper.tar.gz
7z -y e "zookeeper.tar.gz"
7z -y x "zookeeper.tar"
Move-Item zookeeper-3.4.9\* .
New-Item data -type directory
Move-Item conf\zoo_sample.cfg conf\zoo.cfg
"dataDir=C:\ZooKeeper\data" | Out-File conf\zoo.cfg -Encoding ASCII -Append
