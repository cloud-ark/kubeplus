#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    CGO_ENABLED=0 export GOOS=linux; go build .
    #export GOOS=linux; go build .
    cp platform-operator ./artifacts/deployment/platform-operator
    docker build -t gcr.io/cloudark-kubeplus/platform-operator:latest ./artifacts/deployment
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    CGO_ENABLED=0 export GOOS=linux; go build .
    #export GOOS=linux; go build .
    cp platform-operator ./artifacts/deployment/platform-operator
    #echo "PROJECT_ID $PROJECT_ID"
    docker build -t gcr.io/cloudark-kubeplus/platform-operator:$version ./artifacts/deployment
    #docker build -t lmecld/platform-operator:$version ./artifacts/deployment
    #docker push lmecld/platform-operator:$version
    docker push gcr.io/cloudark-kubeplus/platform-operator:$version
fi



