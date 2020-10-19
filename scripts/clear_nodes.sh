#!/bin/bash

cd $PaletteWorkspace

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i"
    rm -rf $node
done
