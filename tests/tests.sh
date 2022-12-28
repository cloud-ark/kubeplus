#!/bin/bash

if (( $# < 1  )); then
	echo "./tests.sh <namespace for KubePlus Operator>"
	exit 0
fi

echo "Running tests.."

minikube_ip=`minikube ip`
if [[ $? != 0 ]]; then
	echo "Minikube not up. Start minikube and retry".
	exit 0
fi

cd ..
KUBEPLUS_NS=$1
export KUBEPLUS_HOME=`pwd`
export PATH=$KUBEPLUS_HOME/plugins:$PATH

echo "Deleting previous resources..."
python3 provider-kubeconfig.py delete $KUBEPLUS_NS

python3 provider-kubeconfig.py create $KUBEPLUS_NS
helm install kubeplus ./deploy/kubeplus-chart -n $KUBEPLUS_NS --kubeconfig=kubeplus-saas-provider.json

until kubectl get pods -n $KUBEPLUS_NS --kubeconfig=kubeplus-saas-provider.json | grep Running; do echo "Waiting for KubePlus Operator Pod to come up..."; sleep 1; done

kubectl create -f tests/resource-quota/wordpress-service-composition.yaml --kubeconfig=kubeplus-saas-provider.json 


until kubectl get crds --kubeconfig=kubeplus-saas-provider.json | grep wordpressservices; do echo "Waiting for WordpressSerivce CRD to register..."; sleep 1; done

kubectl create -f examples/multitenancy/wordpress/tenant2.yaml --kubeconfig=kubeplus-saas-provider.json
kubectl appresources wordpressservices wp-for-tenant3 -k kubeplus-saas-provider.json
kubectl metrics WordpressService wp-for-tenant3 $KUBEPLUS_NS -k kubeplus-saas-provider.json

echo "Press to cleanup test resources.."
read 

kubectl delete -f examples/multitenancy/wordpress/tenant2.yaml --kubeconfig=kubeplus-saas-provider.json 
kubectl delete -f tests/resource-quota/wordpress-service-composition.yaml --kubeconfig=kubeplus-saas-provider.json
helm delete kubeplus -n $KUBEPLUS_NS
echo "Done."
