#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GOOS=linux; go build .
    docker build -t gcr.io/cloudark-kubeplus/grapher:latest .
    docker push gcr.io/cloudark-kubeplus/grapher:latest 
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build -t gcr.io/cloudark-kubeplus/grapher:$version .
    docker push gcr.io/cloudark-kubeplus/grapher:$version
fi



