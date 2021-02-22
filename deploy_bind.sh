#!/bin/bash

make robot t=eth-bind-plt-proxy
make robot t=eth-bind-nft-proxy
make robot t=eth-bind-plt-asset

make robot t=plt-bind-plt-proxy
make robot t=plt-bind-nft-proxy
make robot t=plt-bind-plt-asset
