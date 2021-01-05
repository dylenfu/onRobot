#!/bin/bash

isRemote=$1;
nodeIdx=$2;
currentIp=$3;
sshPort=$4;
identity="node$nodeIdx";

if [[ $isRemote == "false" ]]
then
    kill -INT `ps -ef|grep $identity|grep -v grep|awk '{print $2}'`;
else
#    ssh -p 32000 ubuntu@$currentIp "pid=\$(ps aux | grep '$identity' | awk '{print \$2}' | head -1); echo \$pid |xargs kill";
    ssh -p $sshPort ubuntu@$currentIp "killall -9 geth";
fi

echo "kill $currentIp $identity";