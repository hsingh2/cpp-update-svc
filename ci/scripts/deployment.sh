#!/bin/bash

set -e -x

ROOT=$PWD

if [ ! -d "${ROOT}/build-output/deploy" ]; then
  mkdir -p $ROOT/build-output/deploy
fi

cp resource/cpp-update-svc-* build-output/deploy/cpp-update-svc
cp -r $ROOT/cpp-update-svc-config/$1/* build-output/.
# cp -r $ROOT/cpp-update-svc/instantclient_12_2/ build-output/deploy/instantclient_12_2/

chmod +x $ROOT/build-output/deploy/cpp-update-svc
