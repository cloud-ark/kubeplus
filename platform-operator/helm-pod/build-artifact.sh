#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GO111MODULE=on; CGO_ENABLED=0 export GOOS=linux; go build .
    docker build --no-cache -t gcr.io/cloudark-kubeplus/helm-pod:latest .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GO111MODULE=on; CGO_ENABLED=0 export GOOS=linux; go build .
    docker build -t gcr.io/cloudark-kubeplus/helm-pod:$version .
    docker push gcr.io/cloudark-kubeplus/helm-pod:$version
fi



