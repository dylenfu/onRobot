#!/bin/bash

cd $PaletteWorkspace

killall -INT geth
rm -rf node*

echo "make directions";
./local_mkdir.sh

echo "copy genesis.json and static-nodes.json";
./local_cp_setup_files.sh;

echo "init nodes";
./local_init_nodes.sh

sleep 1s;

echo "start up nodes...";
./local_start_nodes.sh;