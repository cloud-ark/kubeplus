#!/bin/bash

curl https://storage.googleapis.com/kubernetes-helm/helm-v2.11.0-linux-amd64.tar.gz --output helm-v2.11.0-linux-amd64.tar.gz
gunzip helm-v2.11.0-linux-amd64.tar.gz
tar -xvf helm-v2.11.0-linux-amd64.tar


kubectl.sh create serviceaccount --namespace kube-system tiller
kubectl.sh create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
~/kubeplus/linux-amd64/helm init --service-account tiller