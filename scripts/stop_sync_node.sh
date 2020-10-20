#!/bin/bash

cd $PaletteWorkspace

idx=10
node="node$idx"
p2pPort=$(($PaletteStartP2PPort+$idx))

kill -9 `ps -ef|grep geth|grep $p2pPort|grep -v grep|awk '{print $2}'`

rm -rf $node
