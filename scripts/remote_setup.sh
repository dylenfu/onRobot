#!/bin/bash

localWorkspace=$1;
remoteWorkspace=$2;
currentIp=$3;
sshport=$4;

cd $localWorkspace;
rm -rf setup.tar.gz
rm -rf keystore.tar.gz
tar -czvf setup.tar.gz setup
tar -czvf keystore.tar.gz keystore

cmdstr="\
cd $remoteWorkspace;\
cd palette;\
rm -rf setup;\
rm -rf keystore;\
tar -xvf setup.tar.gz;\
tar -xvf keystore.tar.gz;\
rm -rf setup.tar.gz;\
rm -rf keystore.tar.gz;\
";

scp -P $sshport setup.tar.gz ubuntu@${currentIp}:/home/ubuntu/palette/setup.tar.gz;
scp -P $sshport keystore.tar.gz ubuntu@${currentIp}:/home/ubuntu/palette/keystore.tar.gz;
ssh -p $sshport ubuntu@${currentIp} "$cmdstr";