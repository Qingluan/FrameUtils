#!/bin/bash

for i in {1..9}; do
    echo "./start-task-service.sh $i 500$i"
    ./start-task-service.sh $i 500$i

done
