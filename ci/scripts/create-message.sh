#!/usr/bin/env bash

apt-get update && apt-get install -y jq git

VERSION=$(cat version/number)

pushd cpp-update-svc
  SHA=$(git rev-parse HEAD)
  REMOTE=$(git remote get-url origin --push)
popd

jq -n --arg version $VERSION \
  --arg sha $SHA \
  --arg remote $REMOTE \
  '{"deployment": { "version":$version, "artifact_name": "cpp-update-svc", "platform": "R3", "component_id": YOUR-SNS-COMPONENT-ID}}' \
  > message/message.json

cat message/message.json
