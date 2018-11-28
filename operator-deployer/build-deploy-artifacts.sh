#!/bin/bash

rm -f operator-deployer
export GOOS=linux; go build .
cp operator-deployer ./artifacts/deployment/operator-deployer
docker build -t lmecld/operator-deployer:latest ./artifacts/deployment
rm -f artifacts/deployment/operator-deployer

