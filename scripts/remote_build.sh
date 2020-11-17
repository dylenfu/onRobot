#!/bin/bash

currentIp=$1;

cmdstr="\
source /etc/profile;\
cd ~/gohome/src/palette/;\
pwd;\
git pull;\
make geth;\
source /etc/profile;\
geth version;\
";

ssh -p 32000 ubuntu@${currentIp} "$cmdstr";