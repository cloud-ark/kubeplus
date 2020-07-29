## KubePlus - Tooling for Kubernetes native Application Stacks

Kubernetes native stacks are built by extending Kubernetes Resource set (APIs) with Operators and their Custom Resources. Application workflows on Kubernetes are realized by establishing connections between Kubernetes Resources (APIs). These connections can be based on various relationships such as labels, annotations, ownership, etc.

<p align="center">
<img src="./docs/application-workflow.png" width="800" height="300">
</p>

KubePlus tooling simplifies building, visualizing and monitoring these platform workflows. KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

## Summary

KubePlus tooling consists of - kubectl plugins, CRD annotations and (optional) cluster-side add-on.

### CRD Annotations

In Kubernetes application-specific platform workflows are built by establishing relationships between Kubernetes built-in and/or Custom Resources. (e.g. a Service is connected to a Pod through labels.) When working with Custom Resources introduced by Operators, it is important that Operator developer's assumptions around what relationships can be established with a Custom Resource and what actions will be performed as a result of them are clearly articulated. KubePlus provides following annotations on Custom Resource Definitions to encode such assumptions.

```
resource/usage
resource/composition
resource/annotation-relationship
resource/label-relationship
resource/specproperty-relationship
```

More details on how to use these annotations can be found [here](./details.rst). We maintain a table of annotations for Open source Operators that we curate [here](./Operator-annotations.md).

### Client-side kubectl plugins

KubePlus leverages knowledge of relationships between Kubernetes built-in resources and combines that with the CRD annotations mentioned above and builds Kubernetes resource relationship graphs. KubePlus offers a variety of kubectl plugins that internally leverage this graph and enable teams to visualize and monitor platform workflows.

### Cluster-side add-on (optional)

KubePlus also provides an optional PlatformWorkflow Operator that further helps teams define platform workflows that are hard to realize using just helm charts.

## KubePlus kubectl commands

KubePlus offers following kubectl commands (as kubectl plugins)

**1. kubectl man**

- ``kubectl man cr ``: Provides information about how to use a Custom Resource.

**2. kubectl composition**

- ``kubectl composition cr``: Provides information about sub resources created for a Custom Resource instance.

**3. kubectl connections**

- ``kubectl connections cr``: Provides information about relationships of a Custom Resource instance with other resources (custom or built-in) via labels / annotations / spec properties.
- ``kubectl connections service``: Shows all the Pod and Service resources that can be reached from the given service through labels, annotations, or spec properties. 
- ``kubectl connections pod``: Shows all the Service and Pod resources that can be reached from the given pod through labels, annotations, or spec properties.

**4. kubectl metrics**

- ``kubectl metrics cr``: Provides metrics for a Custom Resource instance (count of sub-resources, pods, containers, nodes, total CPU and total Memory consumption).
- ``kubectl metrics service``: Provides CPU/Memory metrics for all the Pods that are descendants of a Service instance. 
- ``kubectl metrics account``: Provides metrics for an account identity - user / service account. (counts of custom resources, built-in workload objects, pods, total CPU and Memory). Needs cluster-side component.
- ``kubectl metrics helmrelease``: Provides CPU/Memory metrics for all the Pods that are part of a Helm release.

**5. kubectl grouplogs**

- ``kubectl grouplogs cr``: Provides logs for all the containers of a Custom Resource instance.
- ``kubectl grouplogs service``: Provides logs for all the containers of all the Pods that are related to a Service object.
- ``kubectl grouplogs helmrelease`` (upcoming): Provides logs for all the containers of all the Pods that are part of a Helm release.


## Example

<p align="center">
<img src="./docs/clusterissuer-mysqlcluster.png" width="750" height="300" class="center">
</p>

``` 
$ kubectl connections service wordpress namespace1

::Final connections graph::
------ Branch 1 ------
Level:0 Service/wordpress
Level:1 Pod/wordpress-pod [related to Service/wordpress by:label]
Level:2 Service/cluster1-mysql-master [related to Pod/wordpress-pod by:envvariable]
Level:3 Pod/cluster1-mysql-0 [related to Service/cluster1-mysql-master by:label]
Level:4 Service/cluster1-mysql-nodes [related to Pod/cluster1-mysql-0 by:envvariable]
Level:4 Service/cluster1-mysql [related to Pod/cluster1-mysql-0 by:label]
Level:4 Service/cluster1-mysql-nodes [related to Pod/cluster1-mysql-0 by:label]
Level:5 MysqlCluster/cluster1 [related to Service/cluster1-mysql-nodes by:owner reference]
Level:6 Service/cluster1-mysql [related to MysqlCluster/cluster1 by:owner reference]
Level:6 Service/cluster1-mysql-master [related to MysqlCluster/cluster1 by:owner reference]
Level:6 ConfigMap/cluster1-mysql [related to MysqlCluster/cluster1 by:owner reference]
Level:6 StatefulSet/cluster1-mysql [related to MysqlCluster/cluster1 by:owner reference]
Level:7 Pod/cluster1-mysql-0 [related to StatefulSet/cluster1-mysql by:owner reference]
------ Branch 2 ------
Level:0 Service/wordpress
Level:1 Ingress/wordpress-ingress [related to Service/wordpress by:specproperty]
Level:2 ClusterIssuer/wordpress-stack [related to Ingress/wordpress-ingress by:annotation]


$ kubectl metrics service wordpress namespace1
---------------------------------------------------------- 
Kubernetes Resources consumed:
    Number of Pods: 2
    Number of Containers: 7
    Number of Nodes: 1
Underlying Physical Resoures consumed:
    Total CPU(cores): 25m
    Total MEMORY(bytes): 307Mi
    Total Storage(bytes): 21Gi
---------------------------------------------------------- 
```

Here is the CRD annotation on the ClusterIssuer Custom Resource:

```
resource/annotation-relationship: on:Ingress, key:cert-manager.io/cluster-issuer, value:INSTANCE.metadata.name
```
The is a annotation-relationship annotation. It defines that Cert Manager looks for 
``cert-manager.io/cluster-issuer`` annotation on Ingress resources. The value of this
annotation is the name of the ClusterIssuer instance.

Here is the CRD annotation on the MysqlCluster Custom Resource:

```
resource/composition: StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget
```

This is the composition annotation. It identifies the set of resources that will be created by MysqlCluster Operator as part of instantiating the MysqlCluster Custom Resource instance.


[Try above example](https://github.com/cloud-ark/kubeplus/blob/master/examples/wordpress-mysqlcluster/steps.txt) in your cluster.

Read [this article](https://medium.com/@cloudark/kubernetes-resource-relationship-graphs-for-application-level-insights-70139e19fb0) to understand more about why tracking resource relationships is useful in Kubernetes.

## Try it:

- To obtain metrics, enable Kubernetes Metrics API Server on your cluster.
  - Hosted Kubernetes solutions like GKE has this already installed.

- KubePlus kubectl commands:

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins-1.0.3.tar.gz
   $ gunzip kubeplus-kubectl-plugins-1.0.3.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins-1.0.3.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

- Cluster-side component:

```
   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
```
- KubePlus kubectl commands:
  - ```$ export KUBEPLUS_HOME=<Full path where kubeplus is cloned>```
  - ```$ export PATH=$KUBEPLUS_HOME/plugins/:$PATH```
- KubePlus Cluster-side add-on:
  - ```cd scripts```
  - ```$ ./deploy-kubeplus.sh```
  - Check out [examples](./examples/kubectl-plugins-and-binding-functions/).

## Operator Maturity Model

In order to build Kubernetes platform workflows using Operators and Custom Resources, it is important for Cluster administrators to evaluate different Operators against a standard set of requirements. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) towards this focusing on Operator usage in multi-Operator environments. We use this model when curating community Operators for enterprise readiness. 


## Status

Actively under development

