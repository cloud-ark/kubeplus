#!/bin/bash

kubectl delete -f artifacts/examples/mystack1.yaml
kubectl delete -f artifacts/examples/platformworkflow-newapi.yaml
kubectl delete -f artifacts/deployment/deployment.yaml
kubectl delete crds mysqlclusterstacks.platformapi.kubeplus
