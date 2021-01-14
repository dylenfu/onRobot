#!/bin/bash

currentIp=$1;
sshport=$2;
remoteGoPath=$3;

cmdstr="\
cd $remoteGoPath/src/palette/;\
source /etc/profile;\
pwd;\
echo '';\
git checkout master;\
git pull origin master;\
git log --pretty=format:'%h - %an, %ar : %s' -2;\
echo '';\
make;\
echo '';\
source /etc/profile;\
echo '';\
geth version;\
";

ssh -p $sshport ubuntu@${currentIp} "$cmdstr";
echo "=========================================";