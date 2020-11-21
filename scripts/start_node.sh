#!/bin/bash

isRemote=$1;
logLevel=$2;
networkID=$3;
currentIp=$4;
nodeIndex=$5;
nodeDir=$6;
rpcPort=$7;
p2pPort=$8;

node="node$nodeIndex";

discflag="--nodiscover --maxpeers 100";
gethflag="--identity $node --datadir data --syncmode full --mine --minerthreads 1 --verbosity $logLevel --networkid $networkID --rpc --rpcaddr 0.0.0.0";
rpcflag="--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints";
runnode="PRIVATE_CONFIG=ignore nohup geth $discflag $gethflag --rpcport $rpcPort $rpcflag --port $p2pPort > node.log 2>&1 &";
cmdstr="\
cd $nodeDir;\
rm -rf node.log;\
rm -rf nohup.out;\
rm -rf data/geth.ipc;\
$runnode";

if [[ $isRemote == "false" ]]
then
    eval $cmdstr;
else
    ssh -p 32000 ubuntu@${currentIp} "\
source /etc/profile;\
$cmdstr";
fi

echo "start $currentIp $node";
sleep 1s;