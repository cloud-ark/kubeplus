#!/bin/bash

#rm -f operator-discovery-helper

version=`tail -1 versions.txt`
export GOOS=linux; go build .
cp operator-discovery-helper ./artifacts/deployment/operator-discovery-helper
docker build -t lmecld/operator-discovery-helper:$version ./artifacts/deployment
#rm -f artifacts/deployment/operator-discovery-helper


