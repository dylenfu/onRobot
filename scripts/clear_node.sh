#!/bin/bash

isRemote=$1;
nodeIdx=$2;
workspace=$3;
currentIp=$4;
node="node$nodeIdx";

cmdstr="\
cd $workspace;\
rm -rf $node;\
";

if [[ $isRemote == "false" ]]; then
    eval $cmdstr;
else
    ssh -p 32000 ubuntu@${currentIp} "$cmdstr";
fi

echo "clear $currentIp $node";