#!/bin/bash

. scp_global.sh

tar -czvf admin.tar.gz admin

# scp set up files
for ip in ${ipList}; do
    echo ${ip};
	scp -P 32000 admin.tar.gz ubuntu@${ip}:/home/ubuntu/palette/admin.tar.gz;
    ssh -p 32000 ubuntu@${ip} "\
    cd palette;\
    rm -rf admin;\
    tar -xvf admin.tar.gz;\
    rm -rf admin.tar.gz";
done

rm -rf admin.tar.gz