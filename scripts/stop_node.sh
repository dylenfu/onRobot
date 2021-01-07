#!/bin/bash

isRemote=$1;
nodeIdx=$2;
currentIp=$3;
sshPort=$4;
identity="node$nodeIdx";

if [[ $isRemote == "false" ]]
then
    kill -s SIGINT $(`ps aux|grep $identity|awk '{print $2}'|head -1`);
else
    ssh -p $sshPort ubuntu@$currentIp "kill -s SIGINT \$(ps aux|grep '$identity' |awk '{print \$2}'|head -1)";
fi

echo "kill $currentIp $identity";
sleep 1s;