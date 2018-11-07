#!/bin/bash

export GOOS=linux; go build .
cp operator-deployer ./artifacts/deployment/operator-deployer
docker build -t lmecld/operator-deployer-test:latest ./artifacts/deployment


