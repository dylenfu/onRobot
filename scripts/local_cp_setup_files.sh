#!/bin/bash

cd $PaletteWorkspace;

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i";
    cp setup/genesis.json $node;
    cp setup/static-nodes.json $node/data/;
    cp setup/$node/nodekey $node/data/geth;
done
