#!/usr/bin/env bash

mkdir -p $GOPATH/bin
CGO_ENABLED=0 go build -o $GOPATH/bin/vamp
