## KubePlus - Tooling for Kubernetes native Application Stacks

Kubernetes native stacks are built by extending Kubernetes clusters with Operators and Custom Resources. In such setups, a DevOps engineer is faced with following challenges:

- How to define and create application-specific platform automation using the Custom and built-in resources available in the cluster?
- How to discover the runtime relationships between different resources (Custom and built-in) in such automation?
- How to troubleshoot issues in such platform automation workflows?
- How to track and correctly attribute physical resource consumption of such workflows at application or team level?


KubePlus solves these issues for DevOps teams. KubePlus tooling simplifies building, visualizing and monitoring Kubernetes platform automation workflows. KubePlus is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code).

KubePlus tooling consists of three components - the Operator Maturity Model for multi-Operator scenarios, client-side kubectl plugins (see below), cluster-side runtime binding resolution component.


## Operator Maturity Model

In order to install and use any Kubernetes Operators in their clusters, Cluster administrators need to evaluate different Operators against a standard set of requirements. Towards this we have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) focusing on Operator usage in multi-Operator environments. Operator developers are using this model today to ensure that their Operator is a good citizen of a multi-Operator world. We use this model when curating community Operators for their multi-Operator readiness. If you are new to Operators, check out [Operator FAQ](https://github.com/cloud-ark/kubeplus/blob/master/Operator-FAQ.md).

## Client-side kubectl plugins

In Kubernetes application-specific platform workflows are built by establishing relationships between Kubernetes built-in and/or Custom Resources. Such relationships can be created using Kubernetes mechanisms of - labels, annotations, Spec properties or owner references. When working with Custom Resources introduced by Operators, it is important that the Operator developer's assumptions around what relationships can be established with a Custom Resource and what actions will be performed as a result of them are clearly articulated. KubePlus provides following annotations to encode such assumptions. These annotations are to be added on a Custom Resource Definition (example below):

```
resource/usage
resource/composition
resource/annotation-relationship
resource/label-relationship
resource/specproperty-relationship
```

More details on how to use these annotations can be found [here](./details.rst). We maintain a table of annotations for Open source Operators that we curate [here](./Operator-annotations.md).

KubePlus leverages knowledge of relationships between Kubernetes built-in resources and combines that with the CRD annotations mentioned above and builds runtime Kubernetes resource topologies. KubePlus offers a variety of kubectl plugins (see below) that internally leverage this topology information and enable DevOps teams to visualize and monitor their platform workflows.


## Cluster-side add-on

For establishing dynamic resource relationships using run time information, KubePlus provides following binding functions. These are resolved by the KubePlus cluster-side add-on using information from the instantiated resources (Custom or built-in) in the cluster.

```
Fn::ImportValue(<Parameter>)
Fn::AddLabel(label, <Resource>)
Fn::AddAnnotation(annotation, <Resource>)
```
Details about binding functions can be found [here](./details.rst#binding-functions). Binding functions are defined in Kubernetes YAMLs. 


## KubePlus kubectl Plugins

KubePlus offers following kubectl commands (as kubectl plugins):

**1. kubectl composition**

- ``kubectl composition``: Provides information about sub resources created for a Kubernetes resource instance.

**2. kubectl connections**

- ``kubectl connections``: Provides information about relationships of a Kubernetes resource instance (custom or built-in) with other resources (custom or built-in) via labels / annotations / spec properties / owner references.

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

- ``kubectl man cr ``: Provides information about how to use a Custom Resource.


## Example

<p align="center">
<img src="./docs/clusterissuer-mysqlcluster.png" width="750" height="300" class="center">
</p>

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

- KubePlus kubectl commands:

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins-latest.tar.gz
   $ gunzip kubeplus-kubectl-plugins-latest.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins-latest.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

- To obtain metrics, enable Kubernetes Metrics API Server on your cluster.
  - Hosted Kubernetes solutions like GKE has this already installed.

- Cluster-side component:

```
   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
   $ cd scripts
   $ ./deploy-kubeplus.sh
```
  - Check out [examples](./examples/kubectl-plugins-and-binding-functions/).


## Status

Actively under development

