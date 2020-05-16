## KubePlus Kubernetes API Add-on

Kubernetes API set is comprised of built-in and Custom Resources. KubePlus API add-on simplifies building platform workflows in Kubernetes YAMLs leveraging these APIs. It offers kubectl plugins that simplify adoption of Custom Resources. It also offers cluster-side component for additional value-add for building and modeling secure and robust platform workflows. 

KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).


## Summary

Kubernetes Custom Resources and Custom Controllers, popularly known as [Operators](https://coreos.com/operators/), extend Kubernetes to run third-party softwares directly on Kubernetes. Teams adopting Kubernetes assemble required Operators of platform softwares such as databases, security, backup etc. to build the required application platforms. KubePlus API add-on simplifies creation of platform level workflows leveraging these Custom Resources.

The primary benefit of using KubePlus to DevOps engineers/Application developers are:

- easily discover static and runtime information about Custom Resources available in their cluster.
- aggregate Custom and built-in resources to build secure and robust platform workflows.


In order to use kubectl commands on Custom Resources, all you need to do is add certain
annotations on the corresponding Custom Resource Definition (CRD) objects. The specific annotations and other details about KubePlus can be found [here](./details.rst).

Even if you are not using Kubernetes Operators or Custom Resources yet, you can 
start using KubePlus kubectl commands with built-in Service resources.


## kubectl commands

KubePlus offers following kubectl commands:

**1. kubectl man**

- ``kubectl man cr``: Provides information about how to use a Custom Resource.

**2. kubectl composition**

- ``kubectl composition cr``: Provides information about sub resources created for a Custom Resource instance.

**3. kubectl connections**

- ``kubectl connections cr``: Provides information about relationships of a Custom Resource instance with other resources (custom or built-in) via labels / annotations / spec properties.
- ``kubectl connections service``: Provides information about relationships between a Service object and all the downstream Pods related to it.

**4. kubectl metrics**

- ``kubectl metrics cr``: Provides metrics for a Custom Resource instance (count of sub-resources, pods, containers, nodes, total CPU and Memory).
- ``kubectl metrics service``: Provides CPU/Memory metrics for all the Pods that are descendants of a Service instance. 
- ``kubectl metrics account``: Provides metrics for an account identity - user / service account. (counts of custom resources, built-in workload objects, pods, total CPU and Memory). Needs on-cluster component.
- ``kubectl metrics helmrelease``: Provides CPU/Memory metrics for all the Pods that are part of a Helm release.

**5. kubectl grouplogs**

- ``kubectl grouplogs cr``: Provides logs for all the containers of a Custom Resource instance.
- ``kubectl grouplogs service``: Provides logs for all the containers of all the Pods that are related to a Service object.
- ``kubectl grouplogs helmrelease`` (upcoming): Provides logs for all the containers of all the Pods that are part of a Helm release.



## Example

``` 
$ kubectl connections service wordpress
Level:1 kind:Pod name:wordpress-6697844b8f-4vlpt relationship-type:label
Level:1 kind:Pod name:wordpress-6697844b8f-8694c relationship-type:label
Level:2 kind:Service name:wordpress-mysql relationship-type:specproperty
Level:3 kind:Pod name:wordpress-mysql-5bf65959f8-w6d25 relationship-type:label

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
```

## Try it:

```
   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
```
- KubePlus kubectl commands:
  - ```$ export PATH=`pwd`/plugins/:$PATH```
  - ```$ export KUBEPLUS_HOME=<Full path where kubeplus is cloned>```
- KubePlus on-cluster component:
  - Use Kubernetes cluster with version 1.14.
  - Enable Kubernetes Metrics API Server on your cluster.
    - Hosted Kubernetes solutions like GKE has this already installed.
  - ```$ ./scripts/deploy-kubeplus.sh```
  - Check out [examples](./examples/moodle-with-presslabs/).


## Operator Maturity Model

In order to build Platform workflows as code using Operators and Custom Resources, it is important for Cluster administrators to evaluate different Operators against a standard set of requirements. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) towards this focusing on Operator usage in multi-Operator environments. We use this model when curating community Operators for enterprise readiness. 


## Operator FAQ

New to Operators? Checkout [Operator FAQ](https://github.com/cloud-ark/kubeplus/blob/master/Operator-FAQ.md).


## Status

Actively under development

