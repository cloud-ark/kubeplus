```markdown
# KubePlus

## Getting Started with an Example

Let’s look at an example of creating a multi-instance WordPress Service using KubePlus. The WordPress service provider goes through the following steps on their cluster.

**NOTE:** If you have not set up KubePlus, follow the [Installation](../README.md#installation) steps to set up KubePlus.

### 1. Create Kubernetes CRD Representing WordPress Helm Chart

*The WordPress Helm chart can be specified as a [public URL](./examples/multitenancy/application-hosting/wordpress/wordpress-service-composition.yaml) or can be [available locally](./examples/multitenancy/application-hosting/wordpress/wordpress-service-composition-localchart.yaml).*

```sh
kubectl create -f https://raw.githubusercontent.com/cloud-ark/kubeplus/master/examples/multitenancy/application-hosting/wordpress/wordpress-service-composition.yaml --kubeconfig=kubeplus-saas-provider.json
kubectl get resourcecompositions
kubectl describe resourcecomposition wordpress-service-composition
```

If the status of the `wordpress-service-composition` indicates that the new CRD has been created successfully, verify it:

```sh
kubectl get crds
```

You should see `wordpressservices.platformapi.kubeplus` CRD registered.

### 2. Create WordpressService Instance `wp-tenant1`

```sh
kubectl create -f https://raw.githubusercontent.com/cloud-ark/kubeplus/master/examples/multitenancy/application-hosting/wordpress/tenant1.yaml --kubeconfig=kubeplus-saas-provider.json
```

### 3. Create WordpressService Instance `wp-tenant2`

```sh
kubectl create -f https://raw.githubusercontent.com/cloud-ark/kubeplus/master/examples/multitenancy/application-hosting/wordpress/tenant2.yaml --kubeconfig=kubeplus-saas-provider.json
```

### 4. Check Created WordpressService Instances

```sh
kubectl get wordpressservices

NAME         AGE
wp-tenant1   86s
wp-tenant2   26s
```

### 5. Check the Details of the Created Instance

```sh
kubectl describe wordpressservices wp-tenant1
```

### 6. Check Created Application Resources

Notice that the `WordpressService` instance resources are deployed in a Namespace `wp-tenant1`, which was created by KubePlus.

```sh
kubectl appresources WordpressService wp-tenant1 -k kubeplus-saas-provider.json

NAMESPACE      KIND                      NAME                      
default        WordpressService          wp-tenant1                
wp-tenant1     PersistentVolumeClaim     mysql-pv-claim            
wp-tenant1     PersistentVolumeClaim     wp-for-tenant1            
wp-tenant1     Service                   wordpress-mysql           
wp-tenant1     Service                   wp-for-tenant1            
wp-tenant1     Deployment                mysql                     
wp-tenant1     Deployment                wp-for-tenant1            
wp-tenant1     Pod                       mysql-76d6d9bdfd-2wl2p    
wp-tenant1     Pod                       wp-for-tenant1-87c4c954-s2cct 
wp-tenant1     NetworkPolicy             allow-external-traffic    
wp-tenant1     NetworkPolicy             restrict-cross-ns-traffic 
wp-tenant1     ResourceQuota             wordpressservice-wp-tenant1
```

### 7. Check Application Resource Consumption

```sh
kubectl metrics WordpressService wp-tenant1 $KUBEPLUS_NS -k kubeplus-saas-provider.json

---------------------------------------------------------- 
Kubernetes Resources created:
    Number of Sub-resources: -
    Number of Pods: 2
        Number of Containers: 2
        Number of Nodes: 1
        Number of Not Running Pods: 0
Underlying Physical Resources consumed:
    Total CPU(cores): 0.773497m
    Total MEMORY(bytes): 516.30859375Mi
    Total Storage(bytes): 40Gi
    Total Network bytes received: 0
    Total Network bytes transferred: 0
---------------------------------------------------------- 
```

### 8. Cleanup

```sh
kubectl delete wordpressservice wp-tenant1 --kubeconfig=kubeplus-saas-provider.json
kubectl delete wordpressservice wp-tenant2 --kubeconfig=kubeplus-saas-provider.json
kubectl delete resourcecomposition wordpress-service-composition --kubeconfig=kubeplus-saas-provider.json
helm delete kubeplus -n $KUBEPLUS_NS
python3 provider-kubeconfig.py delete $KUBEPLUS_NS
```

## Network Isolation Testing

This section verifies that the network policies are correctly isolating application instances.

### Steps

#### Install a Network Driver

On Minikube, install a network driver capable of recognizing `NetworkPolicy` objects (e.g., Cilium):

```sh
minikube start --cni=cilium
eval $(minikube docker-env)
```

#### Refer to Main README for Installing the KubePlus Operator and Plugins

#### Create HelloWorldService Instances

```sh
kubectl create -f hello-world-service-composition.yaml --kubeconfig=provider.conf
kubectl create -f hs1.yaml --kubeconfig=provider.conf
kubectl create -f hs2.yaml --kubeconfig=provider.conf
```

#### Test Network Isolation

- **Ping/HTTP Test from `hs1` to `hs2`:**

  ```sh
  # Get the Pod name for hs1
  HELLOWORLD_POD_HS1=$(kubectl get pods -n hs1 --kubeconfig=provider.conf -o jsonpath='{.items[0].metadata.name}')
  
  # Get the Pod IP for hs2
  HS2_POD_IP=$(kubectl get pods -n hs2 --kubeconfig=provider.conf -o jsonpath='{.items[0].status.podIP}')
  
  # Update and install curl on hs1 pod
  kubectl exec -it $HELLOWORLD_POD_HS1 -n hs1 --kubeconfig=provider.conf -- apt update
  kubectl exec -it $HELLOWORLD_POD_HS1 -n hs1 --kubeconfig=provider.conf -- apt install curl -y
  
  # Test connectivity from hs1 to hs2 using the IP
  kubectl exec -it $HELLOWORLD_POD_HS1 -n hs1 --kubeconfig=provider.conf -- curl $HS2_POD_IP:5000
  ```

  The connection should be denied.

- **Ping/HTTP Test from `hs2` to `hs1`:**

  ```sh
  # Get the Pod name for hs2
  HELLOWORLD_POD_HS2=$(kubectl get pods -n hs2 --kubeconfig=provider.conf -o jsonpath='{.items[0].metadata.name}')
  
  # Get the Pod IP for hs1
  HS1_POD_IP=$(kubectl get pods -n hs1 --kubeconfig=provider.conf -o jsonpath='{.items[0].status.podIP}')
  
  # Update and install curl on hs2 pod
  kubectl exec -it $HELLOWORLD_POD_HS2 -n hs2 --kubeconfig=provider.conf -- apt update
  kubectl exec -it $HELLOWORLD_POD_HS2 -n hs2 --kubeconfig=provider.conf -- apt install curl -y
  
  # Test connectivity from hs2 to hs1 using the IP
  kubectl exec -it $HELLOWORLD_POD_HS2 -n hs2 --kubeconfig=provider.conf -- curl $HS1_POD_IP:5000
  ```

  The connection should be denied.

### Allowing Cross Namespace Traffic

In some scenarios, you might want to enable controlled communication between instances running in different namespaces. KubePlus provides a custom kubectl plugin for this purpose. To allow bi-directional traffic between the two HelloWorldService instances (deployed in namespaces `hs1` and `hs2`), run:

```sh
kubectl allow network traffic hs1 hs2 -k provider.conf
```

```sh 
# Test connectivity from hs1 to hs2 using the IP
kubectl exec -it $HELLOWORLD_POD_HS1 -n hs1 --kubeconfig=provider.conf -- curl $HS2_POD_IP:5000

# Test connectivity from hs2 to hs1 using the IP 
kubectl exec -it $HELLOWORLD_POD_HS2 -n hs2 --kubeconfig=provider.conf -- curl $HS1_POD_IP:5000
```

The connection should be allowed


To deny the traffic between namespace 

```sh
kubectl deny network traffic hs1 hs2 -k provider.conf
```

```sh 
# Test connectivity from hs1 to hs2 using the IP
kubectl exec -it $HELLOWORLD_POD_HS1 -n hs1 --kubeconfig=provider.conf -- curl $HS2_POD_IP:5000

# Test connectivity from hs2 to hs1 using the IP 
kubectl exec -it $HELLOWORLD_POD_HS2 -n hs2 --kubeconfig=provider.conf -- curl $HS1_POD_IP:5000
```

You should see that each policy’s ingress section now includes a rule that uses a namespaceSelector matching the other namespace (using the label `kubernetes.io/metadata.name`).

## Clean Up

```sh
kubectl delete -f hs1-no-replicas.yaml --kubeconfig=provider.conf
kubectl delete -f hs2-no-replicas.yaml --kubeconfig=provider.conf
kubectl delete -f hello-world-service-composition.yaml --kubeconfig=provider.conf
```

Ensure the `helloworldservices.platformapi.kubeplus` CRD is removed.
```

