#!/bin/bash

kubectl delete -f ./kubeplus-components-3.yaml
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding
