#!/bin/bash

rm -f operator-manager
export GOOS=linux; go build .
cp operator-manager ./artifacts/deployment/operator-manager
docker build -t lmecld/operator-manager:latest ./artifacts/deployment
rm -f artifacts/deployment/operator-manager


