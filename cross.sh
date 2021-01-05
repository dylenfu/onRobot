#!/bin/bash

#killall -9 geth
make robot t=clear
make prepare
make compile-dev
make robot t=init,startGenesis,startValidator,addValidators
make robot t=deploy
make robot t=bindProxy,bindAsset,ccmp
make robot t=registerSideChain,approveRegisterSideChain
make robot t=syncGenesis