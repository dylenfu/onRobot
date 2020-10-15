#!/bin/bash

. scp_global.sh

# kill all geth
for ip in ${ipList}; do
    ssh -p 32000 ubuntu@$ip killall -9 geth;
done