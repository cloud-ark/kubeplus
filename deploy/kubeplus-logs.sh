#!/bin/bash

if (( $# < 1 )); then
    echo "./kubeplus-logs.sh <namespace>"
    exit 0
fi

namespace=$1

KUBEPLUS_POD=`kubectl get pods -n $namespace | grep kubeplus | awk '{print $1}'`
#while true; do
  echo "================== Webhook cert setup ===================================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c webhook-cert-setup
  echo "     "
  echo "================== Platform Controller  Logs ===================================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c platform-operator
  echo "     "
  echo "================== Mutating Webhook Logs =========================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c crd-hook
  echo "     "
  echo "================== Helmer Logs ===================================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c helmer
  echo "     "
  echo "================== Consumer UI ===================================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c consumerui
  echo "     "
  echo "================== Mutating Webhook Helper ===================================="
  kubectl logs $KUBEPLUS_POD -n $namespace -c mutating-webhook-helmer
  echo "     "
#  sleep 3
#done
  
