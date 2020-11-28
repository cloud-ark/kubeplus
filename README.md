## KubePlus - Discovery and monitoring of Custom Resources and their dependencies

DevOps teams are using Kubernetes Operators to build custom PaaSes. Kubernetes Operators add Custom Resources to the Kubernetes Resource set. Inventory and chargeback for such a custom PaaS is a challenge today as there is no easy way to discover set of Kubernetes Resources belonging to an application stack. KubePlus addresses this issue by offering generic tooling for discovery and monitoring of Custom Resources and their dependencies.

## What is KubePlus?

KubePlus is a generic tool that enables inventory and chargeback for Kubernetes clusters extended with Operators. It uses a unique method of relationship tags defined on Kubernetes Operator packages (CRDs) to track Kubernetes Custom Resources and their relationships. The tags unlock KubePlus's ability to provide accurate discovery and monitoring for entire Resource set available on the cluster through a set of kubectl plugins.

<p align="center">
<img src="./docs/KubePlus-new.png" width="450" height="300" class="center">
</p>


## Core of KubePlus - Resource Relationship graphs

Operators add Custom Resources (e.g. Mysqlcluster) to the cluster. These resources become first class components of that cluster alongside the built-in resources (e.g. Pod, Service). Platform stacks are realized by establishing relationships between these Kubernetes Resources (built-in or Custom) available on the cluster. These relationships are primarily of four types.
 
(1) Owner references – A resource internally creates additional resources (e.g. MysqlCluster when instantiated, creates Pods and  Services). These sub-resources are related to the parent resource through Owner reference relationship.

(2) Labels and (3) Annotations – Labels or Annotations are key/value pairs that are attached to Kubernetes resources. Resource A can depend on a specific label or annotation to be given on Resource B to take some action.

(4) Spec Properties – Resource A’s Spec property may depend on a value coming from Resource B.

<p align="center">
<img src="./docs/resource-relationship-1.png" width="700" height="300" class="center">
</p>

KubePlus is able to construct Kubernetes Resource relationship graphs like above at runtime. This enables accurate inventory and chargeback tracking for custom PaaSes built using Kubernetes Operators.

## Platform-as-Code tags on CRDs

KubePlus offers following CRD annotations that help Operator developers capture assumptions they have made around what type of relationships can be established with the Custom Resources of their Operators.

```
resource/usage
resource/composition
resource/annotation-relationship
resource/label-relationship
resource/specproperty-relationship
```

Kubernetes Operator developers or cluster administrators can add these annotations to the CRDs. [Here](https://github.com/cloud-ark/kubeplus/blob/master/Operator-annotations.md) are some sample CRD annotations for community Operators that can be used to unlock KubePlus tooling for them.

KubePlus leverages knowledge of relationships between Kubernetes built-in resources and combines that with the CRD annotations mentioned above and builds runtime Kubernetes resource topologies. KubePlus offers a variety of kubectl plugins that internally leverage this topology information.


## Kubectl plugins


**1. kubectl composition**

- ``kubectl composition``: Provides information about sub resources created for a Kubernetes resource instance (custom or built-in). Essentially, 'kubectl composition' shows ownerReference based relationships.

**2. kubectl connections**

- ``kubectl connections``: Provides information about relationships of a Kubernetes resource instance (custom or built-in) with other resources (custom or built-in) via labels, annotations, spec properties and owner references.

**3. kubectl metrics**

- ``kubectl metrics cr``: Provides metrics for a Custom Resource instance (count of sub-resources, pods, containers, nodes, total CPU and total Memory consumption).
- ``kubectl metrics service``: Provides CPU/Memory metrics for all the Pods that are descendants of a Service instance. 
- ``kubectl metrics account``: Provides metrics for an account identity - user / service account. (counts of custom resources, built-in workload objects, pods, total CPU and Memory). Needs cluster-side component.
- ``kubectl metrics helmrelease``: Provides CPU/Memory metrics for all the Pods that are part of a Helm release.

**4. kubectl grouplogs**

- ``kubectl grouplogs cr``: Provides logs for all the containers of a Custom Resource instance.
- ``kubectl grouplogs service``: Provides logs for all the containers of all the Pods that are related to a Service object.
- ``kubectl grouplogs helmrelease`` (upcoming): Provides logs for all the containers of all the Pods that are part of a Helm release.

**5. kubectl man**

- ``kubectl man <Custom Resource> ``: Provides information about how to use a Custom Resource.


## Example

In this example we have two Custom Resources - ClusterIssuer and MysqlCluster. Their CRDs are annotated with following CRD annotations. 

CRD annotation on the ClusterIssuer Custom Resource:

```
resource/annotation-relationship: on:Ingress, key:cert-manager.io/cluster-issuer, value:INSTANCE.metadata.name
```

This defines that CertManager looks for cert-manager.io/cluster-issuer annotation on Ingress resources. The value of this annotation is the name of the ClusterIssuer instance.

CRD annotation on the MysqlCluster Custom Resource:

```
resource/composition: StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget
```

This identifies the set of resources that will be created by the Operator as part of instantiating the MysqlCluster Custom Resource instance. 

<p align="center">
<img src="./docs/clusterissuer-mysqlcluster.png" width="800" height="300" class="center">
</p>

Once these annotations are added to the respective CRDs by the cluster administrator, above resource topology can be discovered by DevOps teams using ``kubectl connections`` plugin as follows:

Note use the kind name as registered with the cluster (e.g.: Deployment or Service and not deployment or service)

``` 
$ kubectl connections Service wordpress namespace1

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
```

The resource consumption of this resource topology can be obtained using ``kubectl metrics`` plugin as follows:

```
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

[Try above example](https://github.com/cloud-ark/kubeplus/blob/master/examples/wordpress-mysqlcluster/steps.txt) in your cluster.

Read [this article](https://medium.com/@cloudark/kubernetes-resource-relationship-graphs-for-application-level-insights-70139e19fb0) to understand more about why tracking resource relationships is useful in Kubernetes.


## Try it:

- KubePlus kubectl commands:

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

- To obtain metrics, enable Kubernetes Metrics API Server on your cluster.
  - Hosted Kubernetes solutions like GKE has this already installed.

  - Check out [examples](./examples/).

## Platform as Code

KubePlus is developed as a part of CloudARK's Platform-as-Code practice. Kubernetes Operators enable extending Kubernetes for application specific workflows. They add Custom Resources and offer foundation for creating application stacks as Code declaratively. Our Platform-as-Code practice offers tools and techniques enabling DevOps teams to build custom PaaSes using Kubernetes Operators.

Platform-as-Code practice consists of:
- Operator Maturity Model:  Operator readiness guidelines for multi-tenant and multi-Operator environment
- KubePlus kubectl plugins: Generic tooling to simplify inventory and charge-back for application stacks created using Operators.
- KubePlus PlatformWorkflow Operator: Publish and monitor Platform Workflows

## Operator Maturity Model

As DevOps team build their custom PaaSes using community or in house developed Operators, they need a set of guidelines for Operator development or evaluation. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) focusing on Operator usage in multi-tenant and multi-Operator environments. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators today for curating community Operators towards building their custom PaaSes.

## PlatformWorkflow Operator

Platform Workflow Operator enables publishing new Services in a cluster. Cluster Admins use this Operator to govern their cluster use by defining and registering opinionated Services with appropriate guard rails. The new Services are registered as new Custom Resources. Application development teams consume the Services by creating instances of these Custom Resources. [Try an example](https://github.com/cloud-ark/kubeplus/tree/master/examples/platform-workflow). 


## Contact

Submit issues on this repository or reach out to our team on [Slack](https://join.slack.com/t/cloudark/shared_invite/zt-2yp5o32u-sOq4ub21TvO_kYgY9ZfFfw).


## Status

Actively under development

