#!/bin/bash

killall -INT geth
source ~/.bash_profile;

sleep 1s;

workspace=${PaletteWorkspace}
cd $workspace;

rm -rf node0/node.log
rm -rf node1/node.log
rm -rf node2/node.log
rm -rf node3/node.log
rm -rf node4/node.log

rm -rf node0/nohup.out
rm -rf node1/nohup.out
rm -rf node2/nohup.out
rm -rf node3/nohup.out
rm -rf node4/nohup.out

rm -rf node0/data/geth.ipc
rm -rf node1/data/geth.ipc
rm -rf node2/data/geth.ipc
rm -rf node3/data/geth.ipc
rm -rf node4/data/geth.ipc

sleep 1s;

gethflag='--datadir data --nodiscover --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0';
rpcflag='--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints';
cd node0
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22000 $rpcflag --port 30300 2>>node.log &

cd ../node1
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22001 $rpcflag --port 30301 2>>node.log &

cd ../node2
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22002 $rpcflag --port 30302 2>>node.log &

cd ../node3
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22003 $rpcflag --port 30303 2>>node.log &

cd ../node4
PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport 22004 $rpcflag --port 30304 2>>node.log &

sleep 1s;

ps -ef|grep geth