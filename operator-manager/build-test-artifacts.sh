#!/bin/bash

export GOOS=linux; go build .
cp operator-manager ./artifacts/deployment/operator-manager
docker build -t lmecld/operator-manager-test:latest ./artifacts/deployment


