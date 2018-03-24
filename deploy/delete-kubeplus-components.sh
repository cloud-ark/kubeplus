#!/bin/bash -x

KUBEPLUS_NS=$1
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding
kubectl delete secret webhook-tls-certificates -n $KUBEPLUS_NS
kubectl delete configmaps kubeplus-saas-consumer-kubeconfig kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS
kubectl delete sa kubeplus-saas-consumer  kubeplus-saas-provider -n $KUBEPLUS_NS
kubectl delete clusterroles kubeplus-saas-consumer kubeplus-saas-provider kubeplus:clusterperms kubeplus:allperms kubeplus:readallperms kubeplus:providerapiperms kubeplus-saas-provider-update
kubectl delete clusterrolebindings kubeplus-saas-consumer kubeplus-saas-provider kubeplus:allperms-binding kubeplus:readallperms-binding kubeplus:providerapiperms-binding kubeplus-saas-provider-update
echo "If you had installed KubePlus using Helm, delete the kubeplus helm release."
