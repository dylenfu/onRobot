#!/bin/bash

cd $PaletteWorkspace

idx=10
node="node$idx"
rpcPort=$(($PaletteStartRPCPort+$idx))
p2pPort=$(($PaletteStartP2PPort+$idx))

kill -9 `ps -ef|grep geth|grep $p2pPort|grep -v grep|awk '{print $2}'`

rm -rf $node
mkdir -p $node/data/geth

# 是否拷贝static-nodes.json无关紧要，因为后面会指定bootnodes
cp setup/genesis.json $node
#cp setup/data/static-nodes.json $node/data/static-nodes.json
cp setup/$node/nodekey $node/data/geth

cd $node
geth --datadir data init genesis.json

PRIVATE_CONFIG=ignore nohup geth \
--datadir data \
--bootnodes  enode://44e509103445d5e8fd290608308d16d08c739655d6994254e413bc1a067838564f7a32ed8fed182450ec2841856c0cc0cd313588a6e25002071596a7363e84b6@127.0.0.1:30300 \
--syncmode full --verbosity 3 \
--networkid $PaletteNetworkID \
--rpc --rpcaddr 127.0.0.1 --rpcport $rpcPort \
--rpcapi db,eth,debug,net,shh,txpool,personal,web3,quorum,istanbul \
--port $p2pPort 2>>node.log &
