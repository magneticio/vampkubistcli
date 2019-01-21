#!/usr/bin/env bash

VERSION=$(go run main.go version clean)

IMAGE=magneticio/vampkubistcli:$VERSION

docker build -t $IMAGE .
docker push $IMAGE

echo $IMAGE
