
Comparison
===========

KubePlus is designed to help platform engineering teams build managed services and multi-tenant SaaS environments on Kubernetes clusters. At a high level it allows you to

1. create your own platform service as a new CRD from a Helm chart. 
2. define policies on this CRD to ensure isolation between the service instances. 
3. get cpu, memory, storage, network consumption metrics for the service instances.
4. get appropriately restricted kubeconfigs for service providers and service consumers.
4. visualize resource relationship graphs to get a snapshot of your service workloads.

This mechanism of attaching policies and consumption metrics to Helm charts enables teams to build their managed services and multi-tenant SaaS in a cookie cutter manner.

For creating new CRDs representing Helm charts there exists [Helm Operator](https://sdk.operatorframework.io/docs/building-operators/helm/tutorial/) from Operator SDK project. In the Helm operator approach, a new Operator is created from scratch per Helm chart. In contrast, KubePlus provides a single ResourceComposition CRD using which new CRDs can be created representing application Helm charts.
ResourceComposition CRD enables defining various policies that are applied to the Helm releases. Another advantage of our approach is that Platform Engineering teams do not have to create a new Operator for every chart. This reduces the number of operator pods running on the cluster if more than one managed services need to be created from different Helm charts.

The aspect of deploying Helm charts in KubePlus is similar to the [Helm controller from the Flux project](https://fluxcd.io/docs/components/helm/). However, there is one crucial difference. KubePlus captures the workflow between a service provider and a service consumer in which creation of a Helm release (i.e. application deployment) is controlled by the service consumer. In contrast, in GitOps workflow, the application deployment is triggered by any updates to the Git repository.

Application-centric components and traits are introduced by the Open Application Model and [KubeVela project](https://github.com/oam-dev/kubevela). The main difference between KubePlus and KubeVela is in our focus on enabling creation of SaaS environments for application providers on Kubernetes, whereas KubeVela is focusing on infrastructure-agnostic application assembly and deployment. This difference is seen in our support for features such as - namespace based multi-instance multi-tenancy, generation of provider and consumer kubeconfig files and tracking resource consumption metrics per Helm release.

KubePlus follows [namespace-based multi-tenancy model](https://kubernetes.io/blog/2021/04/15/three-tenancy-models-for-kubernetes/) where each Helm release is created in a separate namespace. Providers and consumers are granted restricted RBAC permissions that allow them to work with the registered consumer API.
This enables cluster administrators to safely outsource management of the application to the application provider.

For policy enforcement there exist [OPA](https://www.openpolicyagent.org/) and [Kyverno](https://kyverno.io/). These systems are primarily focusing on policy enforcement at cluster / namespace level. As compared to them, KubePlus provides a way to define and enforce policies at Helm chart level thus allowing more fine-grained governance.

For consumption tracking, majority of the monitoring tools show consumption at infrastructure-level of Pods leaving platform teams to develop their own aggregation mechanism based on labels to track consumption at the application-level. With KubePlus you get application-level consumption tracking out of the box as KubePlus tracks cpu, memory, network, storage metrics for each Helm release.

For visualization of resource relationships, there exists [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and various Kubernetes dashboards. Most of these tools are limited to Kubernetes's built-in resources or at best show sub-resource relationship based on owner references. KubePlus tracks all four kinds of resource relationships that can exist between Kubernetes resources (owner references, labels, annotations, spec properties). This enables KubePlus to show  the most complete relationship graphs between the resources in any given namespace. This allows platform engineers to get a sneak peek into what's happening inside a tenant namespace.