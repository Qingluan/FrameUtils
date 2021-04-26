#!/bin/bash
echo $#
[ $# -lt 2 ] &&
	exit 0
sudo docker run --detach \
  --publish $2:4099 \
  --privileged=true \
  --name test-$1 \
  	taskservice
