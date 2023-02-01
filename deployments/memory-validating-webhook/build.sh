#!/bin/bash
cd ../../internal/app/memory-validating-webhook
#: ${DOCKER_USER:? required}

export GO111MODULE=on
export GOPROXY=https://goproxy.cn
# build webhook
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o memory-validating-webhook
# build docker image
docker build --no-cache -t memory-validating-webhook:v1 .
rm -rf memory-validating-webhook

#docker push ${DOCKER_USER}/memory-validating-webhook:v1