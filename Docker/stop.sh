#!/bin/bash

sudo docker ps -a | grep TaskService | awk '{ print $1}' | xargs sudo docker rm -f 

