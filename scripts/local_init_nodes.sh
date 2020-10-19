#!/bin/bash

for((i=$PaletteNodeIndexStart;i<=$PaletteNodeIndexEnd;i++)); do
    nodedir="${PaletteWorkspace}node${i}"
    cd $nodedir
    geth --datadir data init genesis.json
done
