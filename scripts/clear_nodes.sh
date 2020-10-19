#!/bin/bash

cd $PaletteWorkspace

rm -rf node*

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    node="node$i"
    rm -rf $node
done
