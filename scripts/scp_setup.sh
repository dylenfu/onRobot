#!/bin/bash

. scp_global.sh

tar -czvf setup.tar.gz setup

# scp set up files
for ip in ${ipList}; do
    echo ${ip};
	scp -P 32000 setup.tar.gz ubuntu@${ip}:/home/ubuntu/palette/setup.tar.gz;
    ssh -p 32000 ubuntu@${ip} "\
    cd palette;\
    rm -rf setup;\
    tar -xvf setup.tar.gz;\
    rm -rf setup.tar.gz";
done

rm -rf setup.tar.gz