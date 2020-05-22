#!/usr/bin/env bash

set -x

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=on

# Set a default minimum coverage percentage if not set
if [ -z "$MIN_PCT" ]
then
      MIN_PCT=50
fi

cd ${GOPATH}/src/github.comcast.com/cpp/cpp-update-svc
# Gather the list of packages, add vendor if needed
PKG_LIST=$(go list ./... | grep -v fakes | grep -v /vendor/)

# Create the directory for the coverage profiles
[[ -d cover ]] || mkdir cover

# measure the code coverage for each packge
for package in ${PKG_LIST}; do
    go test -tags unit -covermode=count -coverprofile "cover/${package##*/}.cov" "$package" ;
done

# Aggregate the results and calculate the overall percentage
echo "mode: count" > cover/coverage.cover
tail -q -n +2 cover/*.cov >> cover/coverage.cover
total=$( go tool cover -func=cover/coverage.cover | tail -1)
pct=$(echo  $total | cut -d ' ' -f3)

# Exit based on meeting coverage minimum or not.
echo $pct | awk -v min="$MIN_PCT" '$1 >= min { exit 0 } $1 < min { exit 1 }'
