#!/bin/bash

if (( $# < 1 )); then
    echo "./build-artifact.sh <latest | versioned>"
fi

artifacttype=$1

if [ "$artifacttype" = "latest" ]; then
    export GO111MODULE=on; CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crd-hook
    docker build --no-cache -t gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:latest .
fi

if [ "$artifacttype" = "versioned" ]; then
    version=`tail -1 versions.txt`
    echo "Building version $version"
    export GO111MODULE=on; CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crd-hook
    #docker build --no-cache -t lmecld/pac-mutating-admission-webhook:$version .
    #docker push lmecld/pac-mutating-admission-webhook:$version
    docker build --no-cache -t gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:$version .
    docker push gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:$version
fi



