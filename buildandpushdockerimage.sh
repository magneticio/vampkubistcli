#!/usr/bin/env bash

VERSION=$(go run main.go version clean)

IMAGE=magneticio/vamplamiacli:$VERSION

docker build -t $IMAGE .
docker push $IMAGE

echo $IMAGE
