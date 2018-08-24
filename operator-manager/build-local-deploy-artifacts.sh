#!/bin/bash

export GOOS=linux; go build .
cp operator-manager ./artifacts/deployment/operator-manager
docker build -t operator-manager:latest ./artifacts/deployment



