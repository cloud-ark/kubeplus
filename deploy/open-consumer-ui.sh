#!/bin/bash 

KUBEPLUS_POD=`kubectl get pods -A | grep kubeplus | awk '{print $2}'`
KUBEPLUS_NS=`kubectl get pods -A | grep kubeplus | awk '{print $1}'`

kubectl port-forward --address 0.0.0.0 $KUBEPLUS_POD -n $KUBEPLUS_NS 5000:5000 &

echo "Consumer UI available at:"
echo "http://localhost:5000/"



