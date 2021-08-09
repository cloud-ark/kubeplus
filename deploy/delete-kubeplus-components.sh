#!/bin/bash

KUBEPLUS_NS=`kubectl get deployments -A | grep kubeplus-deployment | awk '{print $1}'`
if [[ $KUBEPLUS_NS == '' ]]; then
   KUBEPLUS_NS=default
fi
kubectl delete deployments kubeplus-deployment -n $KUBEPLUS_NS
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding
kubectl delete sa kubeplus -n $KUBEPLUS_NS
kubectl delete svc crd-hook-service -n $KUBEPLUS_NS
kubectl delete svc kubeplus -n $KUBEPLUS_NS
kubectl delete svc kubeplus-consumerui -n $KUBEPLUS_NS
kubectl delete crds resourcecompositions.workflows.kubeplus
kubectl delete crds resourcepolicies.workflows.kubeplus
kubectl delete crds resourceevents.workflows.kubeplus
kubectl delete crds resourcemonitors.workflows.kubeplus
kubectl delete secret webhook-tls-certificates -n $KUBEPLUS_NS
kubectl delete clusterrolebinding kubeplus:cluster-admin
kubectl delete configmaps kubeplus-saas-consumer-kubeconfig kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS
kubectl delete sa kubeplus-saas-consumer  kubeplus-saas-provider -n $KUBEPLUS_NS
kubectl delete clusterroles kubeplus-saas-consumer kubeplus-saas-provider
kubectl delete clusterrolebindings kubeplus-saas-consumer kubeplus-saas-provider
echo "If you had installed KubePlus using Helm, delete the kubeplus helm release."
