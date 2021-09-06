
Comparison
===========

KubePlus is designed to help platform engineering teams build managed services and multi-tenant SaaS environments on Kubernetes clusters. At a high level it allows you to

1. create your own platform service as a new CRD from a Helm chart. 
2. define policies on this CRD to ensure isolation between the service instances. 
3. get cpu, memory, storage, network consumption metrics for the service instances.
4. get appropriately restricted kubeconfigs for SaaS providers and SaaS consumers.
4. visualize resource relationship graphs to get a snapshot of your service workload.

This mechanism of attaching policies and consumption metrics to Helm charts enables teams to build their managed services and multi-tenant SaaS in a cookie cutter manner.

For creating new CRDs representing Helm charts there exists [Helm Operator](https://sdk.operatorframework.io/docs/building-operators/helm/tutorial/) from Operator SDK project. In the Helm operator approach, a new Operator is created from scratch per Helm chart. In contrast, KubePlus provides a single ResourceComposition CRD using which new CRDs can be created representing application Helm charts.
ResourceComposition CRD enables defining various policies that are applied to the Helm releases corresponding to the underlying Helm chart. This aspect allows KubePlus to truly enable platform teams in building managed services or multi-tenant SaaS environments. Another advantage of our approach is that Platform Engineering teams do not have to create a new Operator for every chart. This reduces the number of operator pods running on your clusters if you need to create more than one managed services from different Helm charts.

Application-centric policies and traits are introduced by the Open Application Model and [KubeVela project](https://github.com/oam-dev/kubevela). The main difference between KubePlus and KubeVela is our focus on enabling creation of SaaS environments for application providers, whereas KubeVela is focusing on infrastructure agnostic application deployments. This difference is seen in our features such as - support for namespace based multi-instance multi-tenancy, generation of provider and consumer kubeconfig files and tracking resource consumption per Helm release.

For policy enforcement there exist [OPA](https://www.openpolicyagent.org/) and [Kyverno](https://kyverno.io/). These systems are primarily designed to deliver policy enforcement at cluster / namespace level. As compared to these systems, KubePlus provides a way to define and enforce policies at Helm chart level allowing more fine-grained governance.

For consumption tracking, the majority of the monitoring tools don't show consumption at Custom Resource level, leaving platform teams to develop their own aggregation mechanism based on labels to figure out consumption of underlying resources like cpu, memory, storage, network. With KubePlus you get those details as Prometheus metrics without having to do any other engineering efforts. 

For visualization of resource relationships, there exist [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and various Kubernetes dashboards. Most of these tools are limited to Kubernetes's built-in resources or at the max show sub-resources created by the Custom Resource based on their owner reference relationship. KubePlus is able to track all four kinds of resource relationships that can exist between Kubernetes resources (owner, label, annotation, spec properties) and show you the most complete relationship graphs between the running resources in any given namespace. This allows platform engineers to get a sneak peek into what's happening inside a tenant namespace.