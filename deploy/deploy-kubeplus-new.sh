#!/bin/bash

if (( $# < 1 )); then
    echo "./deploy-kubeplus-new.sh <namespace>"
    exit 0
fi

KUBEPLUS_DEPLOYMENT=`kubectl get deployments -A | grep kubeplus-deployment | awk '{print $1}'`

if [[ -n $KUBEPLUS_DEPLOYMENT ]]; then
   echo "KubePlus already deployed in $KUBEPLUS_DEPLOYMENT namespace. Cannot deploy multiple instances of KubePlus in a cluster."
   echo "Exiting..."
   exit 0
fi

namespace=$1
sed -i .bak s"/namespace:.*/namespace: $namespace/"g kubeplus-components-6.yaml 
kubectl create -f ./kubeplus-components-6.yaml -n $namespace 
