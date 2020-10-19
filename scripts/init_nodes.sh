#!/bin/bash

cd $PaletteWorkspace;

echo "make directions and copy setup files......";
for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i"
    mkdir $node
    mkdir -p $node/data/geth

    cp setup/genesis.json $node;
    cp setup/static-nodes.json $node/data/;
    cp setup/$node/nodekey $node/data/geth;
done

echo "init geth node......";
for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    nodedir="${PaletteWorkspace}node${i}"
    cd $nodedir
    geth --datadir data init genesis.json
done
