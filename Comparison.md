
Comparison
===========

KubePlus belongs to the class of tools that enable `declarative application management`_ in Kubernetes.
As compared to other tools and systems, distinguishing features of KubePlus are - no new CLI, 
focus on Custom Resource stacks, and seamless integration of static and runtime information in realizing such stacks.
In designing KubePlus our main philosophy has been to not introduce any new CLI for enabling
discovery, binding, and orchestration functions.
With KubePlus, application developers use Kubernetes's native CLI 'kubectl' for these functions.
It should be possible though to use 'helm' and/or 'kustomize' with Custom Resource YAMLs defined using KubePlus 
binding functions. 

.. _declarative application management: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/declarative-application-management.md

Problem domain of declarative resource stack creation is not new. In the traditional cloud world,
this problem has been solved by Infrastructure-as-Code tools like AWS CloudFormation and Terraform.
The main assumption that these tools work with is that the set of underlying cloud resource APIs are 
statically known and is not going to change.
With Kubernetes that is not the case. The set of resource APIs available in a cluster
is dynamic as it depends on the Operators/CRDs that are installed in a cluster.
KubePlus API Add-on is solving the declarative platform stack creation problem for this 
dynamic world of Kubernetes CRDs/Operators.

For discovery, Kubernetes itself now supports 'kubectl explain' on Custom Resources.
In our experience the information that is needed for correctly using Custom Resources with other
resources goes beyond the Spec properties that 'kubectl explain' exposes. 
KubePlus's discovery endpoints provide a way for
Operator developers to expose additional information that cannot be accommodated through Custom Resource Spec properties alone.

For orchestration, there exists Application CRD in the community. Conceptually, KubePlus's PlatformStack CRD is
similar to it, in that both provide a way to define a stack of resources.
Our goal with PlatformStack CRD is to use it for orchestration functions such as ordering, label propagation, etc.
Application CRD's focus is mainly on visualization of an application stack.
