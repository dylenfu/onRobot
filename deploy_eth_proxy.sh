#!/bin/bash

make robot t=eth-deploy-plt

make robot t=eth-deploy-plt-proxy
make robot t=eth-plt-ccmp

make robot t=eth-deploy-nft-proxy
make robot t=eth-nft-ccmp
