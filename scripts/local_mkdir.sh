#!/bin/bash

cd $PaletteWorkspace;

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i"
    mkdir $node
    mkdir -p $node/data/geth
done