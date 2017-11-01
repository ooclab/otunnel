#!/bin/bash

./build-by-docker.sh static

TAG=$(./otunnel -v | head -1 |cut -d':' -f 2|tr -d ' \n')
IMAGE_NAME=ooclab/otunnel-amd64:$TAG

docker build -t $IMAGE_NAME -f Dockerfile.scratch .
docker push $IMAGE_NAME
docker tag $IMAGE_NAME ooclab/otunnel-amd64:latest
docker push ooclab/otunnel-amd64:latest
