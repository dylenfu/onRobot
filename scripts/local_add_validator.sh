#!/bin/bash

# how to generate a new nodekey
# istanbul-tools/build/bin/istanbul setup --num 1 --nodes --quorum --verbose
# validators
# {
# 	"Address": "0x6a708455c8777630aac9d1e7702d13f7a865b27c",
# 	"Nodekey": "3d9c828244d3b2da70233a0a2aea7430feda17bded6edd7f0c474163802a431c",
# 	"NodeInfo": "enode://f5135ae0853af71f017a8ecb68e720b729ab92c7123c686e75b7487d4a57ae07dec951380b356246366391ed6cf36f5bcaf39b20c1049ba4a436330406b7b60c@0.0.0.0:30303?discport=0"
# }

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