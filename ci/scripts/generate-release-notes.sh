#!/usr/bin/env bash

set -e

NOTES=$PWD/release-notes

version=$(cat version/number)

echo v${version} > $NOTES/release-name

cat > $NOTES/notes.md <<EOF
See v${version} release notes
EOF

sha=$(git -C cpp-update-svc rev-parse HEAD)
echo $sha > $NOTES/commitish
