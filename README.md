## KubePlus Kubernetes Add-on

KubePlus simplifies building and tracking workflow automation involving Kubernetes Custom Resoures. It brings uniformity in using disparate Custom Resources to create platform workflows. Towards this it provides two things:

- Client-side kubectl commands for discovery, resource usage monitoring and troubleshooting of Custom Resources and their workflows.

- Server-side components for building/modeling secure and robust Custom Resource workflows.

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

- ``kubectl metrics account``: Provides various metrics for an account identity - user / service account. (number of custom resources, number of Deployments/StatefulSets/ReplicaSets/DaemonSets/ReplicationControllers, number of Pods, total CPU and Memory).

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

``` kubectl workflow logs ```

``` kubectl metrics account ```

Details about KubePlus can be found [here](./details.rst). KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

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

- ```kubectl workflow logs Service <service name>```

Check out [examples](./details.rst).


## Status

Actively under development

