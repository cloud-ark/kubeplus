#!/bin/bash

export GOOS=linux; go build .
cp platform-operator ./artifacts/deployment/platform-operator
docker build -t lmecld/platform-operator:0.0.1 ./artifacts/deployment


