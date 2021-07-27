#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    docker build -t gcr.io/cloudark-kubeplus/webhook-tls-getter:latest .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build -t gcr.io/cloudark-kubeplus/webhook-tls-getter:$version .
    docker push gcr.io/cloudark-kubeplus/webhook-tls-getter:$version
fi




