#!/bin/bash

###################################################
#
# prepare on ethereum chain
#
#
###################################################

# deploy eccd, eccm, ccmp contracts and record in config.json
make robot t=eth-deploy-eccd
make robot t=eth-deploy-eccm
make robot t=eth-deploy-ccmp

# transfer eccd ownership to eccm, transfer eccm ownship to ccmp
make robot t=eth-ownership

# register ethereum as an new side chain on poly chain
make robot t=eth-registerSideChain,eth-approveRegisterSideChain

# deploy plt asset
make robot t=eth-deploy-plt

# deploy plt proxy
make robot t=eth-deploy-plt-proxy

# bind plt asset with proxy
make robot t=eth-bind-plt-proxy

# set plt asset in ccmp
make robot t=eth-plt-ccmp

# deploy new nft asset
make robot t=eth-deploy-nft-asset

make robot t=eth-deploy-nft-proxy

make robot t=eth-nft-ccmp

make robot t=eth-bind-nft-proxy
###################################################
#
# prepare on palette chain
#
###################################################
# deploy eccd, eccm, ccmp contracts and record in config.json
make robot t=plt-deploy

# transfer eccd ownership to eccm, transfer eccm ownership to ccmp.
make robot t=plt-ownership

# bind ethereum plt asset and proxy address, set palette ccmp address in palette proxy(PLT contract)
make robot t=plt-bindProxy,plt-bindAsset,plt-ccmp

# register side chain id to poly chain and approve it with 4 poly validators' wallet file.
make robot t=plt-registerSideChain,plt-approveRegisterSideChain

# sync palette header to palette chain and store poly book keepers in the palette chain
make robot t=plt-syncGenesis