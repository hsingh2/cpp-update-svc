#!/bin/bash

set -e -x

ROOT=$PWD

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH

BUILD_DIR=$PWD/resource-swift

VERSION="$(cat version/number)"
ARCH=$(uname)$(uname -m)

cd $GOPATH/src/github.comcast.com/cpp/cpp-update-svc

go build -ldflags "-X github.comcast.com/cpp/cpp-update-svc/server.Versionflag=${VERSION} -X github.comcast.com/cpp/cpp-update-svc/server.Archflag=${ARCH}" -o $BUILD_DIR/cpp-update-svc-$VERSION main.go
