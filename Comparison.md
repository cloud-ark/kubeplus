
Comparison
===========

KubePlus is targetted towards Platform Engineering teams. Broadly it provides four capabilities:

1. Create Kubernetes Custom Resources from Helm charts
2. Define and enforce mutation policies on Custom Resource Pods
3. Visualize Kubernetes resource relationships
4. Monitor Custom Resources

For creating new CRDs from Helm charts there exists [helm-operator](https://docs.okd.io/latest/operators/operator_sdk/osdk-helm.html).
In their approach a new Operator is created from scratch per Helm chart. As compared to that, our approach consists of a single ResourceComposition Operator that enables creating new CRDs from any Helm chart. Advantage of our approach is that Platform Engineering teams do not have to create a new Operator for every chart. 
For policy enforcement there exist [OPA](https://www.openpolicyagent.org/) and [Kyverno](https://kyverno.io/). As compared to these systems, KubePlus provides a way to define and enforce policies at Helm chart level. For visualization there exist [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and various Kubernetes dashboards. These tools primarily show owner reference based relationships and are limited to Kubernetes's built-in resources. KubePlus is able to track all four kinds of relationships that can exist between Kubernetes resources (owner, label, annotation, spec properties). Moreover, KubePlus supports both Kubernetes built-in resources and Custom Resources equally. KubePlus builds resource relationship graphs tracking resources and their relationships. Apart from visualizing relationships between resources, these graphs are useful towards usage monitoring of application stacks consisting of ensemble of Kubernetes resources (built-in and Custom).

The CRD for CRDs technique of KubePlus provides a unique approach to create multi-tenant environments on a cluster. KubePlus supports namespace-based soft multi-tenancy. As compared to other approaches towards this (like [Hierarchical Namespaces](https://kubernetes.io/blog/2020/08/14/introducing-hierarchical-namespaces/)), the CRD for CRDs approach enables us to provide integrated governance and monitoring as part of creating a multi-tenant environment. More generally the CRD for CRDs approach also enables avoiding exchange of Helm charts between Platform Engineering and Product teams.

The approach of registering a Helm chart as part of ResourceComposition CRD has some similarities with various GitOps tools that support Helm. One key difference between those tools and KubePlus is that KubePlus simplifies and formalizes the differences in the workflows that exist today between Platform teams and their cluster users. The ResourceComposition CRD is used by the Platform teams to register a Helm chart as a Kubernetes API (custom resource). The actual deployment of the chart happens when the cluster users create instances of that custom resource.





