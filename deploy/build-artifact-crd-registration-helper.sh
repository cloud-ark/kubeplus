#!/bin/bash

# Exit on error
set -e

if (( $# < 1 )); then
    echo "./build-artifact-crd-registration-helper.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    docker build -f Dockerfile.crd-registration-helper -t gcr.io/cloudark-kubeplus/crd-registration-helper:latest .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    docker build -f Dockerfile.crd-registration-helper -t gcr.io/cloudark-kubeplus/crd-registration-helper:$version .
    docker push gcr.io/cloudark-kubeplus/crd-registration-helper:$version
fi
