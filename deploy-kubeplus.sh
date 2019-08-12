#!/bin/bash

# Start Tiller
helm init

# Deploy KubePlus API Server
kubectl apply -f deploy/

# Deploy KubePlus PlatformStack Operator
kubectl apply -f platform-operator/artifacts/deployment/

# Deploy KubePlus Mutating Webhook
cd mutating-webhook
make deploy
cd ..

# Deploy KubePlus Mutating Webhook helper
kubectl apply -f mutating-webhook-helper/deployment.yaml
