#!/bin/bash

cp setup/genesis.json node0
cp setup/genesis.json node1
cp setup/genesis.json node2
cp setup/genesis.json node3
cp setup/genesis.json node4

cp setup/static-nodes.json node0/data/
cp setup/static-nodes.json node1/data/
cp setup/static-nodes.json node2/data/
cp setup/static-nodes.json node3/data/
cp setup/static-nodes.json node4/data/

cp setup/node0/nodekey node0/data/geth
cp setup/node1/nodekey node1/data/geth
cp setup/node2/nodekey node2/data/geth
cp setup/node3/nodekey node3/data/geth
cp setup/node4/nodekey node4/data/geth