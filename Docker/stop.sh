#!/bin/bash

docker ps -a | grep TaskService | awk '{ print $1}' | xargs docker rm -f 
