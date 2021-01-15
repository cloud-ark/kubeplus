## KubePlus - CRD for CRDs to design Kubernetes platform services from Helm charts

Kubernetes platform engineering teams perform variety of tasks to enable product teams. These tasks include:

- Building deployment workflows offering defaults and abstracting away details
- Attaching logging and monitoring to the application environments
- Controlling quality of service by offering appropriate volumes or backup policies 
- Leveraging policies to ensure appropriate CPU/Memory/Node allocation to Pods

<p align="center">
<img src="./docs/platform-team-challenge.png" width="500" height="250" class="center">
</p>

The key challenge that platform teams face is to offer their services in self-service manner and avoid dreaded exchange of YAMLs between platform and product teams.  To address this challenge, KubePlus offers an open-source framework to create platform services in self-service manner as Kubernetes APIs. It enables platform engineering teams to: 
- Build self-service APIs to embed platform workflows 
- Establish fine-grained control with platform workflow specific policies
- Track CPU/Memory metrics for platform workflows


## KubePlus components

KubePlus has two components: 

### 1. CRD for CRDs to design your platform services from Helm charts

KubePlus offers a CRD named ResourceComposition to 
- Compose new CRDs (Custom Resource Definition) to publish platform services from Helm charts
- Define policies (e.g. Node selection, CPU/Memory limits, etc.) for managing resources of the platform services
- Get aggregated CPU/Memory/Storage Prometheus metrics for the platform services
Here is the high-level structure of ResourceComposition CRD: 

<p align="center">
<img src="./docs/crd-for-crds.png" width="650" height="250" class="center">
</p>

To understand this further let us see how a platform team can build a MySQL service for their product team/s to consume. The base Kubernetes cluster has MySQL Operator on it (either installed by the Platform team or bundled by the Kubernetes provider).

<p align="center">
<img src="./docs/mysql-as-a-service.png" width="400" height="250" class="center">
</p>


The platform workflow requirements are: 
- Create a PersistentVolume of required type for MySQL instance. 
- Create Secret objects for MySQL instance and AWS backup.
- Create MySQL instance with backup target as AWS S3 bucket.  
- Setup a policy in such a way that Pods created under this service will have specified Resource Request and Limits.  
- Get aggregated CPU/Memory metrics for the overall workflow.

Here is a new platform service named MysqlService as Kubernetes API. 

<p align="center">
<img src="./docs/mysql-as-a-service-crd.png" width="650" height="250" class="center">
</p>

A new CRD named MysqlService has been created here using ResourceComposition. You provide a platform workflow Helm chart that creates required underlying resources, and additionally provide policy and monitoring inputs for the workflow. The Spec Properties of MysqlService come from values.yaml of the Helm chart. 
Product teams can use this service to get MySQL database for their application and all the required setups will be performed transparently by this service.


### 2. Kubectl plugins to visualize platform workflows

KubePlus kubectl plugins enable users to discover, monitor and troubleshoot resource relationships in a platform workflow. The plugins run entirely client-side and do not require the in-cluster component. The primary plugin of this functionality is: 
```kubectl connections```: Provides information about relationships of a Kubernetes resource instance (custom or built-in) with other resources (custom or built-in) via owner references, labels, annotations, and spec properties. KubePlus is able to runtime construct Kubernetes Resource relationship graphs. This enables KubePlus to build resource topologies and offer fine grained visibility and control over the platform service.

Here is the resource relationship graph for MysqlSevice created above using 
```kubectl connections MysqlService mysql1'```.

<p align="center">
<img src="./docs/mysqlservice-connections.png" width="750" height="300" class="center">
</p>

We have additional plugins such as ```kubectl metrics``` and ```kubectl grouplogs``` that use resource relationship graphs behind the scene and aggregate metrics and logs for the platform workflow.
You can also directly get CPU/Memory/Storage metrics in Prometheus format if you setup ```ResourceMonitor``` while creating your new CRD. 


## Try it:

- Getting started:
  - Try ```kubectl connections``` plugin. It can be used with any Kubernetes Kind.

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

- CRD for CRDs:
  - Try [this](./examples/resource-composition/steps.txt) example.

- Resource relationship graphs for Public Managed Kubernetes:
  - Check out resource relationship graphs for some public managed Kubernetes providers [here](./examples/graphs). 

- Examples of Operators, Custom Resource stacks, etc.:
  - Check out [examples](./examples/)

- Check [here](./details.rst) for additional details.

Note: To obtain metrics, enable Kubernetes Metrics API Server on your cluster. Hosted Kubernetes solutions like GKE has this already installed.


## Platform-as-Code

KubePlus has been developed as part of our Platform-as-Code practice. Learn more about Platform-as-Code [here](https://cloudark.io/platform-as-code).


## Operator Maturity Model

As enterprise team build their custom PaaSes using community or in house developed Operators, they need a set of guidelines for Operator development and evaluation. We have developed [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) focusing on Operator usage in multi-tenant and multi-Operator environments. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators today for curating community Operators towards building their custom PaaSes.


## Presentations/Talks

1. [Being a good citizen of the Multi-Operator world, Kubecon NA 2020](https://www.youtube.com/watch?v=NEGs0GMJbCw&t=2s)

2. [Operators and Helm: It takes two to Tango, Helm Summit 2019](https://youtu.be/F_Dgz1V5Q2g)

3. [KubePlus presentation at Kubernetes community meeting](https://youtu.be/ZckVULU9sYc)


## Contact

Submit issues on this repository or reach out to our team on [Slack](https://join.slack.com/t/cloudark/shared_invite/zt-2yp5o32u-sOq4ub21TvO_kYgY9ZfFfw).


## Status

Actively under development

