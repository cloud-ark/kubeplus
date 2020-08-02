
Comparison
===========

KubePlus belongs to the class of tools that enable [declarative application management](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md
) in Kubernetes. As compared to other tools, distinguishing features of KubePlus are - 
focus on Custom Resource stacks, seamless integration of static and runtime information in realizing such stacks and no new CLI.

Problem domain of declarative resource stack creation is not new. In the traditional cloud world, this problem has been solved by Infrastructure-as-Code tools like AWS CloudFormation and Terraform. The main assumption that these tools work with is that the underlying cloud resource APIs are statically known and are not going to change.
With Kubernetes that is not the case. The available resource APIs in a cluster
depends on the Operators/CRDs that are installed in that cluster.
KubePlus solves the declarative platform stack creation problem for this 
dynamic world of Kubernetes CRDs/Operators.

For discovery, Kubernetes itself supports 'kubectl explain' on Custom Resources.
In our experience the information that is needed for correctly using Custom Resources alongside other resources goes beyond the Spec properties that 'kubectl explain' exposes. 
KubePlus resource CRD annotations and kubectl plugins provide a way for
Operator developers to expose additional information that cannot be accommodated through Custom Resource Spec properties alone.

For orchestration, there exists Application CRD in the community. Conceptually, KubePlus's PlatformWorkflow CRD is similar to it, in that both provide a way to define a stack of resources. Our goal with PlatformWorkflow CRD is to use it for orchestration functions such as ordering, label propagation, etc. Application CRD's focus is mainly on visualization of an application stack. For visualization we provide client-side kubectl plugins.
