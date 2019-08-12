#!/bin/bash

# Deploy KubePlus API Server
kubectl delete -f deploy/

# Deploy KubePlus PlatformStack Operator
kubectl delete -f platform-operator/artifacts/deployment/

# Deploy KubePlus Mutating Webhook
cd mutating-webhook
make delete
cd ..

# Deploy KubePlus Mutating Webhook helper
kubectl delete -f mutating-webhook-helper/deployment.yaml
