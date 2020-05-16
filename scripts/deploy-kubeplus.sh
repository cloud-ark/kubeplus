#!/bin/bash

# Create Tiller service account
kubectl create -f helm-rbac-config.yaml

# Start Tiller
helm init --service-account tiller

# Deploy KubePlus API Server - Planning to remove KubePlus API Server
# kubectl apply -f ../deploy/

# Deploy KubePlus PlatformStack Operator
kubectl apply -f ../platform-operator/artifacts/deployment/

# Deploy KubePlus Mutating Webhook
cd ../mutating-webhook
make deploy
cd -

