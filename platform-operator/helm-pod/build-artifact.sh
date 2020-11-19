#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GOOS=linux; go build .
    docker build -t gcr.io/disco-horizon-103614/helm-pod:latest .
    docker push gcr.io/disco-horizon-103614/helm-pod:latest
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GOOS=linux; go build .
    docker build -t gcr.io/disco-horizon-103614/helm-pod:version .
    docker push gcr.io/disco-horizon-103614/helm-pod:version
fi



