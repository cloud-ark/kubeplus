## KubePlus Kubernetes API Add-on

Kubernetes API is comprised of built-in and Custom Resources. KubePlus API add-on simplifies building platform workflows in Kubernetes YAMLs leveraging these APIs. It offers kubectl plugins that simplify adoption of Custom Resources. It also offers server side component for additional value-add for building / modeling secure and robust workflows. 

## kubectl commands

KubePlus offers following kubectl commands:

**1. kubectl man**

- ``kubectl man cr``: Provides information about how to use a Custom Resource.

**2. kubectl composition**

- ``kubectl composition cr``: Provides information about sub resources created as part of handling a Custom Resource instance.

**3. kubectl connections**

- ``kubectl connections cr``(upcoming): Provides information about relationships of a Custom Resource instance with other resources (custom or built-in) via labels / annotations / spec properties / sub-resources.

- ``kubectl connections workflow``: Provides information about relationships between a Service object and all the downstream Pods related to it representing a workflow.

**4. kubectl metrics**

- ``kubectl metrics cr``: Provides various metrics for Custom Resource instance (number of sub-resources, number of pods, number of containers, number of nodes on which the pods run, total CPU and Memory).

- ``kubectl metrics account``: Provides various metrics for an account identity - user / service account. (number of custom resources, number of Deployments/StatefulSets/ReplicaSets/DaemonSets/ReplicationControllers, number of Pods, total CPU and Memory). Needs server-side component.

- ``kubectl metrics workflow`` (upcoming): Provides CPU/Memory metrics for all the Pods that are part of a Workflow (direct and indirect descendants).

**5. kubectl grouplogs**

- ``kubectl grouplogs cr``:Provides logs for all the containers of a Custom Resource instance.

- ``kubectl grouplogs workflow``: Provides logs for all the containers of all the Pods that are part of the workflow defined by the provided Service instance.


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

If you are not yet using Operators or Custom Resources, you can still use following commands:

``` kubectl connections workflow ```

``` kubectl grouplogs workflow ```

Details about KubePlus can be found [here](./details.rst). KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).


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


## Quick details

In order to use KubePlus all you need to do is enhance Custom Resource Definition (CRD) YAMLs with following annotations.

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



## Status

Actively under development

