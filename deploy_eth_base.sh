#!/bin/bash

make robot t=eth-deploy-eccd
make robot t=eth-deploy-eccm
make robot t=eth-deploy-ccmp

make robot t=eth-eccd-ownership
make robot t=eth-eccm-ownership

make robot t=eth-registerSideChain
make robot t=eth-approveRegisterSideChain

make robot t=eth-syncGenesis