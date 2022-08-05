#!/bin/bash -x

namespace=$1
python3 /root/kubeconfiggenerator.py $namespace
kubectl label --overwrite=true ns $namespace managedby=kubeplus

# Just be around
while [ True ]
do
   sleep 1000
done
