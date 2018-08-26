#!/bin/bash

export GOOS=linux; go build .
cp operator-deployer ./artifacts/deployment/operator-deployer
docker build -t operator-deployer:latest ./artifacts/deployment



