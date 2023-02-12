#!/usr/bin/env bash

LDFLAGS=${LDFLAGS-"-s -w"}

# set -x

for os in "linux" "darwin"; do
    for arch in "amd64" "arm64"; do
        echo "$os/$arch: building"
        GOOS=$os GOARCH=$arch go build -ldflags="$LDFLAGS" -o "bin/passthroughtools-server-$os-$arch"
        echo "$os/$arch: done"
    done
done
