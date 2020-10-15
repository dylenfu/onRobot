#!/bin/bash

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