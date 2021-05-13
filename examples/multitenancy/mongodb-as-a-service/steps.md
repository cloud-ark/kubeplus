## Setup

- Install Docker 
- Install Helm v3
- Create a cluster
- Install KubePlus kubectl plugins
- Install KubePlus server-side component


## Deploy Mongodb Operator

- Install Mongodb Operator on the cluster
  ```helm install mongodboperator https://github.com/cloud-ark/operatorcharts/blob/master/mongodb-0.1.0.tgz?raw=true```


## Deploy mongodbservicetenant
```
kubectl apply -f mongodbservicetenant.yaml
```

## Wait for the new service to get created

```
kubectl get crds | grep mongodbservicetenants.platformapi.kubeplus
```

## Deploy tenant1
```
kubectl apply -f tenant1.yaml
```

## Check composition
```
kubectl connections MongoDBServiceTenant tenant1 default -o png -i ServiceAccount:default
```

## Check policies
```
kubectl get pods example-mongodb-0 -n namespace1 -o json | jq -r '.spec.containers[0].resources'
```

## Check monitoring
### open terminal 1
```
KUBEPLUS_POD=`kubectl get pods | grep kubeplus | awk '{print $1}'`
kubectl port-forward $KUBEPLUS_POD -n default 8081:8090
```
### open terminal 2
```
curl -kv "http://127.0.0.1:8081/apis/kubeplus/metrics?kind=MongoDBServiceTenant&instance=tenant1&namespace=default"
```  

### Cleanup

```kubectl delete -f mongodbservicetenant.yaml ```
