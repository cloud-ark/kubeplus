#!/bin/bash

bash ./webhook-create-signed-cert.sh --service crd-hook-service --namespace default --secret webhook-tls-certificates
cat ./mutatingwebhook.yaml | ./webhook-patch-ca-bundle.sh > ./mutatingwebhook-ca-bundle.yaml
kubectl apply -f ./mutatingwebhook-ca-bundle.yaml
kubectl apply -f ./kubeplus-components-3.yaml
