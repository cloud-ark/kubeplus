#!/bin/bash

KUBEPLUS_NS=`kubectl get deployments -A | grep kubeplus-deployment | awk '{print $1}'`
kubectl delete deployments kubeplus-deployment -n $KUBEPLUS_NS
kubectl delete mutatingwebhookconfigurations platform-as-code.crd-binding
kubectl delete sa kubeplus -n $KUBEPLUS_NS
kubectl delete svc crd-hook-service -n $KUBEPLUS_NS
kubectl delete svc kubeplus -n $KUBEPLUS_NS
kubectl delete crds resourcecompositions.workflows.kubeplus
kubectl delete crds resourcepolicies.workflows.kubeplus
kubectl delete crds resourceevents.workflows.kubeplus
kubectl delete crds resourcemonitors.workflows.kubeplus
kubectl delete secret webhook-tls-certificates -n $KUBEPLUS_NS
kubectl delete clusterrolebinding kubeplus:cluster-admin



