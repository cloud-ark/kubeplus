#!/bin/bash

if (( $# < 1 )); then
    echo "./deploy-kubeplus-new.sh <namespace>"
    exit 0
fi

namespace=$1
sed -i .bak s"/namespace:.*/namespace: $namespace/"g kubeplus-components-6.yaml 
kubectl create -f ./kubeplus-components-6.yaml -n $namespace 
