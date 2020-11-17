#!/bin/bash

isRemote=$1;
currentIp=$2;

echo "is remote=$isRemote, current ip=$currentIp";
cmdstr="ps -ef|grep geth";

if [[ $isRemote = "false" ]]
then
    eval $cmdstr;
else
    ssh -p 32000 ubuntu@${currentIp} "$cmdstr";
fi