#!/usr/bin/env bash

VERSION=0.0.5

IMAGE=magneticio/vamplamiacli:$VERSION

docker build -t $IMAGE .
docker push $IMAGE

echo $IMAGE
