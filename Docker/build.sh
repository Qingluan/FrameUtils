#!/bin/bash
go build  -ldflags="-s -w" ../services/TaskService/
go build  -ldflags="-s -w" ../console/NetTest/
sudo docker build -f  Dockerfile  -t taskservice:latest .

rm TaskService
