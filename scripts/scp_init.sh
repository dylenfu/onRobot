#!/bin/bash

. scp_global.sh

# node0 ~ node3 in ip1~ip4
for((i=1;i<=$maxNodes;i++)); do
echo "$currentIp $currentNode";

ssh -p 32000 ubuntu@${currentIp} "\
    cd palette;\
    source /etc/profile;\
    rm -rf $currentNode;\
    mkdir -p $currentNode/data/geth;\
    mkdir -p $currentNode/data/keystore;\
    cp setup/$currentNode/nodekey $currentNode/data/geth/;\
    cp setup/genesis.json $currentNode/;\
    cp setup/scp-nodes.json $currentNode/data/static-nodes.json;\
    cp -r admin/keystore/* $currentNode/data/keystore/;\
    cd $currentNode;\
    geth --datadir data init genesis.json;\
    ";

nodeIncrease;
done
