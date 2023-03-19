#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1
cd ../../

if [ "$artifacttype" = "latest" ]; then
    export GO111MODULE=on; export GOOS=linux; go build .
    docker build --no-cache -t gcr.io/cloudark-kubeplus/helm-pod:latest  -f ./platform-operator/helm-pod/Dockerfile .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GO111MODULE=on; export GOOS=linux; go build .
    docker build -t gcr.io/cloudark-kubeplus/helm-pod:$version ./platform-operator/helm-pod/Dockerfile .
    docker push gcr.io/cloudark-kubeplus/helm-pod:$version
fi



