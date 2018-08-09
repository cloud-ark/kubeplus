#!/bin/bash

export GOOS=linux; go build .
cp postgres-crd-v2 ./artifacts/deployment/postgres-crd-v2
#docker build -t postgres-crd-v2:latest ./artifacts/deployment
docker build -t lmecld/postgres-crd-v2:latest ./artifacts/deployment


