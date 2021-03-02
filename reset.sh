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