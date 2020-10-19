#!/bin/bash

rm -rf node5
mkdir -p node5/data/geth

cp setup/genesis.json node5
cp setup/static-nodes.json node5/data/
cp setup/node5/nodekey node5/data/geth

cd node5
geth --datadir data init genesis.json

gethflag='--datadir data --nodiscover --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0';
rpcflag='--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints';
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22005 $rpcflag --port 30305 2>>node.log &

ps -ef|grep geth