## KubePlus API Add-on

KubePlus API add-on simplifies building workflow automation involving Kubernetes Custom Resoures. It brings uniformity in using disparate Custom Resources to create platform workflows. Towards this it provides kubectl commands for discovery, resource usage monitoring and troubleshooting of Custom Resources. 

## kubectl commands

KubePlus API add-on offers following kubectl commands:

**1. kubectl man <Custom Resource>:** Provides information about how to use a Custom Resource.

**2. kubectl tree <Custom Resource Instance>:** Provides information about sub resources created as part of handling a Custom Resource instance.

**3. kubectl graph <Custom Resource Instance>: (upcoming)** Provides information about relationships of a Custom Resource instance with other resources (custom or built-in) via labels / annotations / spec properties / sub-resources.

**4. kubectl metrics cr <Custom Resource Instance>:** Provides various metrics for Custom Resource instance (number of sub-resources, number of pods, number of containers, number of nodes on which the pods run, total CPU and Memory).

**5. kubectl metrics account <Account Name>:** Provides various metrics for an account identity - user / service account. (number of custom resources, number of Deployments/StatefulSets/ReplicaSets/DaemonSets/ReplicationControllers, number of Pods, total CPU and Memory).

**6. kubectl crlogs <Custom Resource Instance>:** Provides logs for all the containers of a Custom Resource instance.

## Example

``` 
$ kubectl metrics cr MysqlCluster cluster1 namespace1
---------------------------------------------------------- 
 Creator Account Identity: devdattakulkarni@gmail.com
---------------------------------------------------------- 
 Number of Sub-resources: 7
 Number of Pods: 2
 Number of Containers: 16
 Number of Nodes: 1
Total CPU(cores): 84m
Total MEMORY(bytes): 302Mi
----------------------------------------------------------

$ kubectl metrics account devdattakulkarni@gmail.com
---------------------------------------------------------- 
 Creator Account Identity: devdattakulkarni@gmail.com
---------------------------------------------------------- 
 Number of Custom Resources: 3
 Number of Deployments: 1
 Number of StatefulSets: 0
 Number of ReplicaSets: 0
 Number of DaemonSets: 0
 Number of ReplicationControllers: 0
 Number of Pods: 0
Total CPU(cores): 288m
Total MEMORY(bytes): 524Mi
```

Details about KubePlus API Add-on can be found [here](./details.rst). KubePlus API Add-on is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

## How to use?

In order to use KubePlus API add-on all you need to do is enhance Custom Resource Definition (CRD) YAMLs with following annotations.

```
platform-as-code/usage
```

The 'usage' annotation is used to define usage information for a Custom Resource.
The value for 'usage' annotation is the name of the ConfigMap that stores the usage information.

```
platform-as-code/composition
```

The 'composition' annotation is used to define Kubernetes's built-in resources that are created as part of instantiating a Custom Resource instance. 

Here is an example of MysqlCluster Custom Resource Definition (CRD) enhanced with above annotations:

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: mysqlclusters.mysql.presslabs.org
    annotations:
      helm.sh/hook: crd-install
      platform-as-code/usage: mysqlcluster-usage.usage
      platform-as-code/composition: StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget
  spec:
    group: mysql.presslabs.org
    names:
      kind: MysqlCluster
      plural: mysqlclusters
      shortNames:
      - mysql
    scope: Namespaced
```

## Try it:

- Use Kubernetes cluster with version 1.14.
- Enable Kubernetes Metrics API Server on your cluster.
  - Hosted Kubernetes solutions like GKE has this already installed.

```
   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
   $ ./deploy-kubeplus.sh
   $ export PATH=`pwd`:$PATH
```

Check out [examples](./details.rst).


## Status

Actively under development

