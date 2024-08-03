#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact-cleanup.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    docker build -t gcr.io/cloudark-kubeplus/delete-kubeplus-resources:latest -f ./Dockerfile.cleanup .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build -t gcr.io/cloudark-kubeplus/delete-kubeplus-resources:$version -f ./Dockerfile.cleanup .
    docker push gcr.io/cloudark-kubeplus/delete-kubeplus-resources:$version
fi




