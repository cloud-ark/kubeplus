#!/bin/bash

KUBEPLUS_NS=`oc get deployments -A | grep kubeplus-deployment | awk '{print $1}'`
oc delete deployments kubeplus-deployment -n $KUBEPLUS_NS
oc delete mutatingwebhookconfigurations platform-as-code.crd-binding
oc delete sa kubeplus -n $KUBEPLUS_NS
oc delete svc crd-hook-service -n $KUBEPLUS_NS
oc delete svc kubeplus -n $KUBEPLUS_NS
oc delete crds resourcecompositions.workflows.kubeplus
oc delete crds resourcepolicies.workflows.kubeplus
oc delete crds resourceevents.workflows.kubeplus
oc delete crds resourcemonitors.workflows.kubeplus
oc delete secret webhook-tls-certificates -n $KUBEPLUS_NS
oc delete clusterrolebinding kubeplus:cluster-admin
