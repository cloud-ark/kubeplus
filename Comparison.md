
Comparison
===========

KubePlus is targetted towards Platform Engineering teams. Broadly it provides four capabilities:
1. Create Kubernetes Custom Resources from Helm charts
2. Define and enforce mutation policies on Custom Resource Pods
3. Visualize Kubernetes resource relationships
4. Monitor Custom Resources

For creating new CRDs from Helm charts there exists [helm-operator](https://docs.okd.io/latest/operators/operator_sdk/osdk-helm.html).
In their approach a new Operator is created from scratch per Helm chart. As compared to that, our approach consists of a single ResourceComposition Operator that enables creating new CRDs from any Helm chart. Advantage of our approach is that Platform Engineering teams do not have to create a new Operator for every chart. 
For policy enforcement there exist [OPA](https://www.openpolicyagent.org/) and [Kyverno](https://kyverno.io/). Distinguishing feature of KubePlus is our focus on Pod spec mutations for Pods belonging to Custom Resource instances. For visualization there exist [kubectl-tree](https://github.com/ahmetb/kubectl-tree) and various Kubernetes dashboards. These tools primarily show owner reference based relationships and are limited to Kubernetes's built-in resources. KubePlus is able to track all four kinds of relationships that can exist between Kubernetes resources (owner, label, annotation, spec properties). Moreover, KubePlus supports both Kubernetes built-in resources and Custom Resources equally. KubePlus builds resource relationship graphs tracking resources and their relationships. Apart from visualizing relationships between resources, these graphs are useful towards usage monitoring of application stacks consisting of ensemble of Kubernetes resources (built-in and Custom).