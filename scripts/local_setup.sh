#!/bin/bash

mkdir -p setup
cd setup
../istanbul-tools/build/bin/istanbul setup --num 5 --nodes --quorum --save --verbose