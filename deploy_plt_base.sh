#!/bin/bash

make robot t=plt-deploy-eccd
make robot t=plt-deploy-eccm
make robot t=plt-deploy-ccmp

make robot t=plt-eccd-ownership
make robot t=plt-eccm-ownership

make robot t=plt-plt-ccmp
make robot t=plt-deploy-nft-proxy
make robot t=plt-nft-ccmp

make robot t=plt-registerSideChain
make robot t=plt-approveRegisterSideChain
make robot t=plt-syncGenesis