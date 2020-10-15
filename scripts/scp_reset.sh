#!/bin/bash

# stop all nodes
./scp_stop_all.sh

# clear all nodes
./scp_clear_all.sh

# init
./scp_init.sh

# run
./scp_start_all.sh