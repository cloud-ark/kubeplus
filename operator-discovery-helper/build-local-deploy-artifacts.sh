#!/bin/bash

export GOOS=linux; go build .
cp operator-discovery-helper ./artifacts/deployment/operator-discovery-helper
docker build -t operator-discovery-helper:latest ./artifacts/deployment



