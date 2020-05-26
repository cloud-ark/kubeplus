## KubePlus - Tooling for Kubernetes native Application Stacks

One of the key reasons for Kubernetes’s popularity is its extensibility. Kubernetes API extensions (commonly referred as [Operators](https://coreos.com/operators/)) extend Kubernetes API and enable adding application specific workflow automation in Kubernetes-native manner. There are a wide variety of Operators built today for softwares like databases, key-value stores, API gateways etc. to run on Kubernetes. Enterprise DevOps teams assemble required Kubernetes Operators and create their Kubernetes-native application stacks. The key challenge when working with such stacks is the need to easily discover and use the Custom APIs/Resources available in a cluster towards creating required application-specific workflows. 

KubePlus consists of suite of tools that simplify building, visualizing and monitoring Kubernetes application workflows that are made up of Kubernetes's built-in and Custom Resources available in a cluster.

You can start using KubePlus by simply annotating your Custom Resource Definitions (CRDs) with certain annotations (outlined below). 
If you are not yet using Kubernetes Operators or Custom Resources, 
you can still use KubePlus to visualize and monitor application workflows that are made up of Kubernetes's built-in resources such as Services and Pods.

KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

## Summary

KubePlus tool suite consists of following components - CRD Annotations, client-side kubectl plugins, and an optional in-cluster component.

### CRD Annotations

Application workflows are built by establishing relationships between Kubernetes built-in and/or Custom Resources. (e.g. a Service is connected to a Pod through labels.) Kubernetes offers labels, annotations and spec properties to define resource relationships. When working with Custom Resources introduced by Operators, it is important that Operator developer's assumptions around what relationships can be established with a Custom Resource and what actions will be performed as a result of them are clearly articulated. KubePlus provides a set of annotations on Custom Resource Definitions to encode such assumptions. The specific annotations and how to use them can be found [here](./details.rst))

### Client-side kubectl plugins

KubePlus leverages knowledge of relationships between Kubernetes built-in resources and combines that with the CRD annotations mentioned above and builds Kubernetes resource relationship graph. KubePlus offers number of kubectl plugins that internally leverages this graph and enables teams to visualize and monitor application workflows.  

### In-cluster component (optional)

KubePlus also provides an optional PlatformWorkflow Operator that further helps teams define application workflows that are hard to realize using just helm charts.

## KubePlus kubectl commands

KubePlus offers following kubectl commands (as kubectl plugins)

**1. kubectl man**

- ``kubectl man cr``: Provides information about how to use a Custom Resource.

**2. kubectl composition**

- ``kubectl composition cr``: Provides information about sub resources created for a Custom Resource instance.

**3. kubectl connections**

- ``kubectl connections cr``: Provides information about relationships of a Custom Resource instance with other resources (custom or built-in) via labels / annotations / spec properties.
- ``kubectl connections service``: Shows all the Pod and Service resources that can be reached from the given service through labels, annotations, or spec properties. 
- ``kubectl connections pod``: Shows all the Service and Pod resources that can be reached from the given pod through labels, annotations, or spec properties.

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
Level:0 kind:Service name:wordpress Owner:/
Level:1 kind:Pod name:wordpress-6697844b8f-7m627 Owner:Deployment/wordpress
Level:1 kind:Pod name:wordpress-6697844b8f-kx7wg Owner:Deployment/wordpress
Level:2 kind:Service name:wordpress-mysql Owner:/
Level:3 kind:Pod name:wordpress-mysql-5bf65959f8-fmxpx Owner:Deployment/wordpress-mysql


$ kubectl connections pod wordpress-mysql-5bf65959f8-fmxpx default 
Level:0 kind:Pod name:wordpress-mysql-5bf65959f8-fmxpx Owner:Deployment/wordpress-mysql
Level:1 kind:Service name:wordpress-mysql Owner:/
Level:2 kind:Pod name:wordpress-6697844b8f-7m627 Owner:Deployment/wordpress
Level:2 kind:Pod name:wordpress-6697844b8f-kx7wg Owner:Deployment/wordpress
Level:3 kind:Service name:wordpress Owner:/


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
- To obtain metrics, enable Kubernetes Metrics API Server on your cluster.
  - Hosted Kubernetes solutions like GKE has this already installed.
- KubePlus kubectl commands:
  - ```$ export KUBEPLUS_HOME=<Full path where kubeplus is cloned>```
  - ```$ export PATH=`pwd`/plugins/:$PATH```
- KubePlus In-cluster component:
  - ```$ ./scripts/deploy-kubeplus.sh```
  - Check out [examples](./examples/moodle-with-presslabs/).

## Operator Maturity Model

In order to build Kubernetes application workflows using Operators and Custom Resources, it is important for Cluster administrators to evaluate different Operators against a standard set of requirements. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) towards this focusing on Operator usage in multi-Operator environments. We use this model when curating community Operators for enterprise readiness. 


## Status

Actively under development

