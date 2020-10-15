#!/bin/bash

. scp_global.sh

# kill all geth
for ip in ${ipList}; do
    ssh -p 32000 ubuntu@$ip killall -9 geth;
done

sleep 1s;

# todo rm -rf geth.ipc
for((i=1;i<=$maxNodes;i++)); do
echo "$currentIp $currentNode $currentPort $currentRPCPort";

cmdstr="PRIVATE_CONFIG=ignore \
nohup geth --datadir data --nodiscover \
--istanbul.blockperiod 15 --syncmode full \
--mine --minerthreads 1 --verbosity 5 \
--networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport $currentRPCPort \
--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul \
--emitcheckpoints --port $currentPort \
> node.log 2>&1 &";

ssh -p 32000 ubuntu@${currentIp} "\
    cd palette;\
    cd $currentNode;\
    source /etc/profile;\
    rm -rf node.log;\
    rm -rf nohup.out;\
    $cmdstr";

nodeIncrease;
sleep 1s;
done

for ip in ${ipList}; do
    echo "${ip} node number:"
	ssh -p 32000 ubuntu@${ip} "ps -ef|grep geth|grep -v grep|wc -l"
done