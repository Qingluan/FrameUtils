#!/bin/bash

if [[ $1 == "kibana" ]]; then

docker run -d \
    --name ki \
    --link es:elasticsearch \
    -p 5601:5601 \
    kibana
#    -e "elasticsearch.url=localhost:9200" \
fi

[[ $1 == "es" ]] &&
docker run --name es -p 9200:9200 -p 9300:9300  -e "discovery.type=single-node" -e ES_JAVA_OPTS="-Xms320m -Xmx640m" -d elasticsearch


[[ $1 == "rm" ]] &&
docker rm -f es

