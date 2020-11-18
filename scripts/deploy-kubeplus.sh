#!/bin/bash

cd ../deploy

# Create Tiller service account
kubectl create -f helm-rbac-config.yaml

# Start Tiller
helm init --service-account tiller

# Deploy KubePlus API Server - Planning to remove KubePlus API Server
# kubectl apply -f ../deploy/

# Deploy KubePlus PlatformStack Operator
# kubectl apply -f ../platform-operator/artifacts/deployment/

# Deploy KubePlus Mutating Webhook
#cd ../mutating-webhook
#make deploy
#cd -

# Deploy Mutating Webhook helper
#kubectl apply -f ../deploy/kubeplus-components-1.yaml

bash ./webhook-create-signed-cert.sh --service crd-hook-service --namespace default --secret webhook-tls-certificates
cat ./mutatingwebhook.yaml | ./webhook-patch-ca-bundle.sh > ./mutatingwebhook-ca-bundle.yaml
kubectl apply -f ./mutatingwebhook-ca-bundle.yaml
kubectl apply -f ./kubeplus-components-2.yaml