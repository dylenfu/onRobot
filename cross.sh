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
make robot t=eth-registerSideChain
make robot t=eth-approveRegisterSideChain

# deploy plt asset
make robot t=eth-deploy-plt

# deploy plt proxy
make robot t=eth-deploy-plt-proxy
make robot t=eth-plt-ccmp

make robot t=eth-deploy-nft-proxy
make robot t=eth-nft-ccmp

make robot t=eth-syncGenesis

###################################################
#
# prepare on palette chain
#
###################################################
# deploy eccd, eccm, ccmp contracts and record in config.json
make robot t=plt-deploy

# transfer eccd ownership to eccm, transfer eccm ownership to ccmp.
make robot t=plt-ownership

# register side chain id to poly chain and approve it with 4 poly validators' wallet file.
make robot t=plt-registerSideChain,plt-approveRegisterSideChain

make robot t=plt-ccmp

# sync palette header to palette chain and store poly book keepers in the palette chain
make robot t=plt-syncGenesis

###################################################
#
# bind proxy and asset on both of palette and ethereum
#
###################################################
# bind plt asset and proxy
make robot t=eth-bind-plt-proxy
make robot t=eth-bind-plt-asset
make robot t=eth-bind-nft-proxy
make robot t=eth-deploy-nft-asset
make robot t=eth-bind-nft-asset

# bind ethereum plt asset and proxy address, set palette ccmp address in palette proxy(PLT contract)
make robot t=plt-bindProxy
make robot t=plt-bindAsset
