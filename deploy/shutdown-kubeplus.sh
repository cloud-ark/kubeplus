#!/bin/bash

# Delete KubePlus API Server
# kubectl delete -f ../deploy/

# Delete KubePlus PlatformStack Operator
# kubectl delete -f ../platform-operator/artifacts/deployment/

#kubectl delete -f helm-rbac-config.yaml

#kubectl delete deployments tiller-deploy -n kube-system

# Delete KubePlus Mutating Webhook
#cd ../mutating-webhook
#make delete
#cd -

# Delete KubePlus Mutating Webhook helper
kubectl delete -f ./kubeplus-components-2.yaml
