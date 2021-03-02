#!/bin/bash

make robot t=stop
make prepare
make compile
make robot t=clear,init,start
./deploy_eth_base.sh
./deploy_eth_proxy.sh
./deploy_plt_base.sh
./deploy_bind.sh
make robot t=eth-plt-mint-gov
make robot t=eth-plt-mint-admin
./transfer_plt_ownership.sh

make robot t=eth-eth-transfer
make robot t=plt-lock,plt-unlock

make robot t=plt-deploy-nft-asset
make robot t=eth-deploy-nft-asset

# 修改BindNFTAsset.json以及NFT-Lock.json&NFT-UnLock.json
make robot t=plt-bind-nft-asset
make robot t=eth-bind-nft-asset

make robot t=nft-lock,nft-unlock