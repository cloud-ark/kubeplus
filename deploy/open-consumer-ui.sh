#!/bin/bash 

if (( $# < 1 )); then
    echo "./open-consumerui.sh <consumer kubeconfig>"
    exit 0
fi

consumerkubecfg=$1

KUBEPLUS_POD=`kubectl get pods -A --kubeconfig=$consumerkubecfg | grep kubeplus | awk '{print $2}'`
KUBEPLUS_NS=`kubectl get pods -A --kubeconfig=$consumerkubecfg | grep kubeplus | awk '{print $1}'`

kubectl port-forward --address 0.0.0.0 $KUBEPLUS_POD -n $KUBEPLUS_NS 5000:5000 --kubeconfig=$consumerkubecfg &

echo "Consumer UI available at:"
echo "http://localhost:5000/"
echo "-----------------------"
echo "Create service instances by using ffollowing URL:"
echo "http://localhost:5000/service/<service name>"
echo "Example:"
echo "http://localhost:5000/service/HelloWorldService"



