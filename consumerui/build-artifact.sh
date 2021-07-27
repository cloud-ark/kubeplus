#!/bin/bash

if (( $# < 1 )); then
    echo "./build-consumerui.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    docker build --no-cache -t gcr.io/cloudark-kubeplus/consumerui:latest .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build --no-cache -t gcr.io/cloudark-kubeplus/consumerui:$version .
    docker push gcr.io/cloudark-kubeplus/consumerui:$version
fi




