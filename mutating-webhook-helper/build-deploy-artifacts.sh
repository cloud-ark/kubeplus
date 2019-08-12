#!/bin/bash

export GOOS=linux; go build .
docker build -t lmecld/mutating-webhook-helper:latest .


