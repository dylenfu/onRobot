#!/bin/bash

killall -INT geth

rm -rf node0 node1 node2 node3 node4

echo "make directions";
./local_mkdir.sh

echo "copy genesis.json and static-nodes.json";
./local_cp_setup_files.sh;

echo "init node1";
cd node0
geth --datadir data init genesis.json

cd ../node1/
geth --datadir data init genesis.json

cd ../node2/
geth --datadir data init genesis.json

cd ../node3/
geth --datadir data init genesis.json

cd ../node4/
geth --datadir data init genesis.json

sleep 1s;

echo "start up nodes...";
cd ../;
./local_start_nodes.sh;