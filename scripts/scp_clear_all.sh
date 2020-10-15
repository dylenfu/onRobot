#!/bin/bash

. scp_global.sh

for ip in ${ipList}; do
    ssh -p 32000 ubuntu@${ip} "\
    cd palette;\
    rm -rf node*;";
done