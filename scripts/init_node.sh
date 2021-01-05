#!/bin/bash

nodeIdx=$1;
isRemote=$2;
workspace=$3;
currentIp=$4;
sshport=$5;

node="node$nodeIdx";

cmdstr="\
cd $workspace;\
mkdir $node;\
mkdir -p $node/data/geth;\
\
cp setup/genesis.json $node/;\
cp setup/static-nodes.json $node/data/;\
cp setup/$node/nodekey $node/data/geth;\
cd $node;\
geth --datadir data init genesis.json;\
";

if [[ $isRemote == "false" ]]
then
    eval $cmdstr;
else
    ssh -p $sshport ubuntu@${currentIp} "\
source /etc/profile;\
$cmdstr";
fi
