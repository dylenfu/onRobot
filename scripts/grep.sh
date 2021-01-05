#!/bin/bash

isRemote=$1;
currentIp=$2;
sshPort=$3;

echo "is remote=$isRemote, current ip=$currentIp";
cmdstr="ps -ef|grep geth";

if [[ $isRemote = "false" ]]
then
    eval $cmdstr;
else
    ssh -p $sshPort ubuntu@${currentIp} "$cmdstr";
fi