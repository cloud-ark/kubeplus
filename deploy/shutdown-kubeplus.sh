#!/bin/bash

if (( $# < 1 )); then
    echo "./shutdown-kubeplus.sh <namespace>"
    exit 0
fi

namespace=$1

kubectl delete -f ./kubeplus-components-6.yaml -n $namespace
kubectl delete -f ./kubeplus-non-pod-resources.yaml -n $namespace
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding -n $namespace
kubectl delete serviceaccount kubeplus -n $namespace
