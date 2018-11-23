#!/usr/bin/env bash

# don't forget to export GOPATH

echo using GOPATH as $GOPATH

docker run --rm -it -v "$GOPATH":/go -w /go/src/github.com/magneticio/vamp2cli dockercore/golang-cross:1.11.1 sh -c '
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    if [ "$GOOS" = "windows" ]; then
      go get -u github.com/spf13/cobra
    fi
    go build -o bin/vamp2cli-$GOOS-$GOARCH
  done
done
'
echo "Binaries can be found under bin directory"
