#!/bin/bash

isRemote=$1;
nodeIdx=$2;
currentIp=$3;
sshPort=$4;
identity="node$nodeIdx";

if [[ $isRemote == "false" ]]
then
    kill -s SIGINT $(ps aux|grep geth|grep $identity|grep -v grep|awk '{print $2}');
else
    ssh -p $sshPort ubuntu@$currentIp "kill -s SIGINT \$(ps aux|grep geth|grep '$identity'|grep -v grep|awk '{print \$2}')";
fi

echo "kill $currentIp $identity";
sleep 1s;