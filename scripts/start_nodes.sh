#!/bin/bash

gethflag="--datadir data --nodiscover --syncmode full --mine --minerthreads 1 --verbosity $PaletteLogLevel --networkid $PaletteNetworkID --rpc --rpcaddr 127.0.0.1";
rpcflag="--rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints";

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i"
    rm -rf $node/node.log
    rm -rf $node/nohup.out
    rm -rf $node/data/geth.ipc

    nodedir="${PaletteWorkspace}${node}"
    cd $nodedir

    rpcPort=$(($PaletteStartRPCPort+$i))
    p2pPort=$(($PaletteStartP2PPort+$i))
    PRIVATE_CONFIG=ignore nohup geth $gethflag --rpcport $rpcPort $rpcflag --port $p2pPort 2>>node.log &
    sleep 1s;
done

ps -ef|grep geth