#!/bin/bash

currentIp=$1;
sshport=$2;
remoteGoPath=$3;

cmdstr="\
killall -9 geth;\
cd $remoteGoPath/src/palette/;\
pwd;\
git checkout master;\
git pull origin master;\
make geth;\
source /etc/profile;\
geth version;\
";

ssh -p $sshport ubuntu@${currentIp} "$cmdstr";