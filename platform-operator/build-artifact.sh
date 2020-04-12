#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GOOS=linux; go build .
    cp platform-operator ./artifacts/deployment/platform-operator
    docker build -t lmecld/platform-operator:latest ./artifacts/deployment
    docker push lmecld/platform-operator:latest
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GOOS=linux; go build .
    cp platform-operator ./artifacts/deployment/platform-operator
    docker build -t lmecld/platform-operator:$version ./artifacts/deployment
    docker push lmecld/platform-operator:$version
fi



