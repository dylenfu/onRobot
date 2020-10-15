#!/bin/bash

workspace=/Users/dylen/software/palette-chain-example

# set ip list
ipList="\
106.75.246.130 \
106.75.250.160 \
106.75.251.68 \
106.75.232.131 \
"

IFS=' ' read -ra ipArray <<< "$ipList"

ipLength=${#ipArray[@]};

maxNodes=5;

# set global node name
currentNodeIdx=0;
currentNode="node$currentNodeIdx";
currentIp=${ipArray[0]};
startPort=30300;
startRPCPort=22000;
currentPort=$startPort;
currentRPCPort=$startRPCPort;

nodeIncrease() {
currentNodeIdx=$((currentNodeIdx+1));
currentNode="node${currentNodeIdx}";
ipmod=$(($currentNodeIdx%$ipLength));
currentIp=${ipArray[$ipmod]};
portmod=$(($currentNodeIdx/$ipLength));
currentPort=`expr $startPort + $portmod`;
currentRPCPort=`expr $startRPCPort + $portmod`;
}
