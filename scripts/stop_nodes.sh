#!/bin/bash

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    p2pPort=$(($PaletteStartP2PPort+$i))
    echo "kill node$i, p2p port $p2pPort"
    kill `ps -ef|grep geth|grep $p2pPort|grep -v grep|awk '{print $2}'`
done

ps -ef|grep geth