#!/usr/bin/env bash

set -ex

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=on

pushd ${GOPATH}/src/github.comcast.com/cpp/cpp-update-svc
  go test ./... -v -tags unit
popd
