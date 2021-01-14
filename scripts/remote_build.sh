#!/bin/bash

currentIp=$1;
sshport=$2;
remoteGoPath=$3;

cmdstr="\
cd $remoteGoPath/src/palette/;\
source /etc/profile;\
pwd;\
git checkout master;\
git pull origin master;\
git log -p -1;\
make;\
source /etc/profile;\
geth version;\
";

ssh -p $sshport ubuntu@${currentIp} "$cmdstr";
echo "=========================================";