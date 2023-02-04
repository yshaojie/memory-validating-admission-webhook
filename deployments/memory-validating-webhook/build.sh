#!/bin/bash
cd ../../cmd/memory-validating-webhook
#: ${DOCKER_USER:? required}

export GO111MODULE=on
export GOPROXY=https://goproxy.cn
# build webhook
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o memory-validating-webhook
# build docker image
docker build --no-cache -t yshaojie/memory-validating-webhook:v1 .
rm -rf memory-validating-webhook
#如果采用kind启动k8s集群，则需要把本地镜像导入到k8s节点中
#具体可以参考https://www.lixueduan.com/posts/kubernetes/15-kind-kubernetes-in-docker/
kind load docker-image yshaojie/memory-validating-webhook:v1 --name kind
#docker push ${DOCKER_USER}/memory-validating-webhook:v1