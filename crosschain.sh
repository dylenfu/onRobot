#!/bin/bash  

num=100
make compile env=test

for((i=1;i<=$num;i++));
do
make robot t=nft-lock,nft-unlock
done