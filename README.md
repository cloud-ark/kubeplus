## KubePlus - Tooling for Kubernetes native Application Workflows

One of the key reasons for Kubernetes’s popularity is its extensibility. Kubernetes API extensions (commonly referred as [Operators](https://coreos.com/operators/)) extend Kubernetes API and enable adding application specific workflow automation in Kubernetes-native manner. There are a wide variety of Operators built today for softwares like databases, key-value stores, API gateways etc. to run on Kubernetes. Enterprise DevOps teams assemble required Kubernetes Operators and create their Kubernetes-native application stacks. The key challenge when working with such stacks is the need to easily discover and use the available Custom APIs/Resources in a cluster towards creating required application-specific workflows. 

KubePlus consists of suite of tools that simplify building, visualizing and monitoring Kubernetes-native application workflows that are made up of Kubernetes's built-in and Custom Resources available in a cluster.

You can start using KubePlus by simply annotating your Custom Resource Definitions (CRDs) with certain annotations (outlined below). 
If you are not yet using Kubernetes Operators or Custom Resources, 
you can still use KubePlus to visualize and monitor application workflows that are made up of Kubernetes's built-in resources such as Services and Pods.

KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

## Summary

KubePlus tool suite consists of:

### Resource Annotations

At its core, the Kubernetes resource model is built around the notion of relationships. Kubernetes offers following mechanisms to define resource relationships - labels with label selectors, annotations and spec properties. Kubernetes-native application workflows are built by establishing relationships between built-in and/or Custom Resources. For instance, a Service is connected to a Pod through labels and label selectors. When working with Operators and Custom Resources, it is important that such relationships be easy to discover, define and use.
KubePlus provides a set of annotations to encode such relationships in a standard manner.
The specific annotations and how to use them can be found [here](./details.rst))

### Client-side plugins

Typically, a Kubernetes-native application workflow is identified using one of the following three things: a top-level Service resource, a Custom Resource instance, or a Helm release. 
KubePlus provides client-side kubectl plugins to visualize and monitor Kubernetes-native application workflows. 

### Cluster-side component

KubePlus provides server-side PlatformWorkflow Operator to define workflows involving multiple Custom Resources that depend on the sub-resources created by their respective Operators.


### Operator Maturity Model

In order to build Kubernetes-native application workflows using Operators and Custom Resources, it is important for Cluster administrators to evaluate different Operators against a standard set of requirements. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) towards this focusing on Operator usage in multi-Operator environments. We use this model when curating community Operators for enterprise readiness. 


## KubePlus kubectl commands

KubePlus offers following kubectl commands (as kubectl plugins)

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
  - ```$ export KUBEPLUS_HOME=<Full path where kubeplus is cloned>```
  - ```$ export PATH=`pwd`/plugins/:$PATH```
- KubePlus cluster-side component:
  - Use Kubernetes cluster with version 1.14.
  - Enable Kubernetes Metrics API Server on your cluster.
    - Hosted Kubernetes solutions like GKE has this already installed.
  - ```$ ./scripts/deploy-kubeplus.sh```
  - Check out [examples](./examples/moodle-with-presslabs/).


## Status

Actively under development

