#!/bin/bash

cd $PaletteWorkspace

./init_nodes.sh
sleep 1s;

echo "start up nodes...";
./start_nodes.sh;