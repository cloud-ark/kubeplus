#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact-kubeconfiggenerator.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    docker build -f  ./Dockerfile.kubeconfiggenerator -t gcr.io/cloudark-kubeplus/kubeconfiggenerator:latest . 
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build -f  ./Dockerfile.kubeconfiggenerator -t gcr.io/cloudark-kubeplus/kubeconfiggenerator:$version .
    docker push gcr.io/cloudark-kubeplus/kubeconfiggenerator:$version
fi




