#!/bin/bash

rm -f operator-discovery-helper
export GOOS=linux; go build .
cp operator-discovery-helper ./artifacts/deployment/operator-discovery-helper
docker build -t lmecld/operator-discovery-helper:latest ./artifacts/deployment
rm -f artifacts/deployment/operator-discovery-helper


