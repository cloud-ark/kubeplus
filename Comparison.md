
Comparison
===========

KubePlus is designed to help platform engineering teams build managed services or multi-tenant SaaS environments on Kubernetes clusters. At a high level it allows you to

1. create your own platform service as a new CRD from a Helm chart. 
2. define policies on this CRD to ensure isolation between the service instances. 
3. get cpu/memory/storage consumption metrics for the service instances. 
4. visualize resource relationship graphs to get a snapshot of your service workload.

This mechanism of attaching policies and consumption metrics to the Helm chart enables teams to build their managed services or multitenant SaaS in a cookie cutter model. 

For creating new CRDs from Helm charts there exists [helm-operator](https://docs.okd.io/latest/operators/operator_sdk/osdk-helm.html). In the helm-operator approach, a new Operator is created from scratch per Helm chart. As compared to that, our approach consists of a single ResourceComposition Operator that enables creating new CRDs on the fly from any Helm chart. Advantage of our approach is that Platform Engineering teams do not have to create a new Operator for every chart. So it not only reduces the admin efforts in building CRDs but also significantly reduces the number of operator pods running on your clusters if you need to create more than one managed services from Helm charts. However, the primary difference is in our ability to attach policies and prometheus monitoring to these CRDs. This aspect allows KubePlus to truly enable platform teams in building managed services or multi-tenant SaaS. 

For policy enforcement there exist [OPA](https://www.openpolicyagent.org/) and [Kyverno](https://kyverno.io/). These systems are primarily designed to deliver policy enforcement at cluster / namespace level. As compared to these systems, KubePlus provides a way to define and enforce policies at Helm chart level allowing more fine-grained governance. 

For consumption tracking, the majority of the monitoring tools don't show consumption at Custom Resource level, leaving platform teams to develop their own clumsy aggregation mechanism based on labels to figure out consumption of underlying resources like CPU, memory, storage. With KubePlus you get those details as Prometheus metrics without having to do any other engineering efforts. 

For visualization of resource relationships, there exist [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and various Kubernetes dashboards. Most of these tools are limited to Kubernetes's built-in resources or at the max show sub-resources created by the Custom Resource based on their owner reference relationship. KubePlus is able to track all four kinds of resource relationships that can exist between Kubernetes resources (owner, label, annotation, spec properties) and show you the most complete relationship graphs between the running resources in any given namespace. This allows platform engineers to get a sneak peek into what's happening inside a tenant namespace. 

Thus the CRD for CRDs technique of KubePlus provides a unique and comprehensive approach to create your managed services or multi-tenant SaaS. Platform engineering teams are now equipped with a cookie cutter approach to build their services without getting into unmanageable exchange of helm charts with their product teams. 
