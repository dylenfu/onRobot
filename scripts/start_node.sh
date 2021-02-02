#!/bin/bash

isRemote=$1;
logLevel=$2;
networkID=$3;
currentIp=$4;
nodeIndex=$5;
nodeDir=$6;
rpcPort=$7;
p2pPort=$8;
sshPort=$9;

node="node$nodeIndex";

discflag="--nodiscover --maxpeers 100";
#gcflag="--gcmode=archive";
gethflag="--identity $node --datadir data --syncmode full --mine --minerthreads 1 --verbosity $logLevel --networkid $networkID --rpc --rpcaddr 0.0.0.0";
rpcflag="--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints";
runnode="PRIVATE_CONFIG=ignore nohup geth $discflag $gethflag --rpcport $rpcPort $rpcflag --port $p2pPort > node.log 2>&1 &";
cmdstr="\
cd $nodeDir;\
pwd; \
rm -rf geth.ipc; \
ls -al; \
rm -rf nohup.out;\
source /etc/profile;\
$runnode";

if [[ $isRemote == "false" ]]
then
    eval $cmdstr;
else
    ssh -p $sshPort ubuntu@${currentIp} "$cmdstr";
fi

echo "start $currentIp $node";
sleep 1s;