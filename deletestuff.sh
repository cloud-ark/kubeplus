#!/bin/bash

kubectl delete deployments kubeplus-deployment postgres-operator-deployment
kubectl delete configmap postgres-crd-v2-chart

kubectl delete customresourcedefinition postgreses.postgrescontroller.kubeplus 

helm list | awk '{print $1}' | grep -v NAME | xargs helm delete

