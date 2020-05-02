#!/bin/bash

# Delete KubePlus API Server
kubectl delete -f ../deploy/

# Delete KubePlus PlatformStack Operator
kubectl delete -f ../platform-operator/artifacts/deployment/

# Delete KubePlus Mutating Webhook
cd ../mutating-webhook
make delete
cd -

# Delete KubePlus Mutating Webhook helper
#kubectl delete -f mutating-webhook-helper/deployment.yaml
