# vamp lamia command line client

## development
if you have go installed,
git clone it to $GOPATH/src/github.com/magneticio/vamp2cli
so that docker builder works.

## build

for docker build:
```
./build.sh
```
for local build:
```
./build.sh local
```

binaries will be under bin directory

## Run
For mac run:
```
./bin/vamp2cli-darwin-amd64 --help
```

Copy the binary for you platform to the user binaries folder for general usage, for MacOS:

```
cp vamp2cli-darwin-amd64 /usr/local/bin/vamp2cli
chmod +x /usr/local/bin/vamp2cli
```
