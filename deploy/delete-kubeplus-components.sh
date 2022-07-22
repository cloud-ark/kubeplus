#!/bin/bash -x

KUBEPLUS_NS=$1
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding
kubectl delete crds resourcecompositions.workflows.kubeplus
kubectl delete crds resourcepolicies.workflows.kubeplus
kubectl delete crds resourceevents.workflows.kubeplus
kubectl delete crds resourcemonitors.workflows.kubeplus
kubectl delete secret webhook-tls-certificates -n $KUBEPLUS_NS
kubectl delete configmaps kubeplus-saas-consumer-kubeconfig kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS
kubectl delete sa kubeplus-saas-consumer  kubeplus-saas-provider -n $KUBEPLUS_NS
kubectl delete clusterroles kubeplus-saas-consumer kubeplus-saas-provider
kubectl delete clusterrolebindings kubeplus-saas-consumer kubeplus-saas-provider
echo "If you had installed KubePlus using Helm, delete the kubeplus helm release."
