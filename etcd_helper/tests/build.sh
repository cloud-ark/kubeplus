#!/bin/bash

export GOOS=linux;
cd ..
go test -c ./tests -o etcdhelper.test
cd tests
mv ../etcdhelper.test ./
docker build -t etcd-helper-tests:latest ./
