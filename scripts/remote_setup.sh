#!/bin/bash

localWorkspace=$1;
remoteWorkspace=$2;
currentIp=$3;

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

scp -P 32000 setup.tar.gz ubuntu@${currentIp}:/home/ubuntu/palette/setup.tar.gz;
scp -P 32000 keystore.tar.gz ubuntu@${currentIp}:/home/ubuntu/palette/keystore.tar.gz;
ssh -p 32000 ubuntu@${currentIp} "$cmdstr";