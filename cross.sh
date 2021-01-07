#!/bin/bash

# deploy eccd, eccm, ccmp contracts and record in config.json
# and transfer eccd ownership to eccm, transfer eccm ownership to ccmp.
make robot t=deploy,ownership

# bind ethereum plt asset and proxy address, set palette ccmp address in palette proxy(PLT contract)
make robot t=bindProxy,bindAsset,ccmp

# todo: poly dail failed
# register side chain id to poly chain and approve it with 4 poly validators' wallet file.
make robot t=registerSideChain,approveRegisterSideChain

# sync palette header to palette chain and store poly book keepers in the palette chain
make robot t=syncGenesis