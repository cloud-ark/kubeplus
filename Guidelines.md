# Kubernetes Operator guidelines for Multi-tenancy

A Kubernetes Operator adds new Custom Resources into a cluster towards managing applications like databases, key-value stores, API gateways, etc. in Kubernetes native manner. An Operator enables creating multiple instances of the application that they are managing. This makes them a key building block in enabling SaaS based delivery of the applications. Below are design aspects that Operator developers should keep in mind to ensure that their Operator is ready for such multi-instance multi-tenancy SaaS use-case.

[1. Namespace based multi-tenancy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#namespace-based-multi-tenancy)

[2. Atomicity of application instance deployment](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#atomicity-of-application-instance-deployment)

[3. Segregate application instances on different worker nodes](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#segregate-application-instances-on-different-worker-nodes)

[4. Co-location of application Pods on a worker node](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#co-location-of-application-pods-on-a-worker-node)

[5. Prevent cross-traffic between application instances](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#prevent-cross-traffic-between-application-instances)


# Detail Guidelines


### Namespace based multi-tenancy

To support multi-tenant use-cases, your Operator should support creating resources within different namespaces rather than just in the namespace
in which it is deployed (or the default namespace). This allows your Operator to support namespace based multi-tenancy.


### Atomicity of application instance deployment

Kubernetes provides mechanism of [requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-types) for specifying the cpu and memory resource needs of a Pod's containers. When specified, Kubernetes scheduler ensures that the Pod is scheduled on a Node that has enough capacity 
for these resources. A Pod with request and limits specified for every container is given ``guaranteed`` Quality-of-Service (QoS) by the Kubernetes scheduler. A Pod in which only resource requests are specified for at least one container is given ``burstable`` QoS. A Pod with no requests/limits specified is given ``best effort`` QoS.
It is important to define resource requests and limits for your Custom Resources.

This helps with ensuring ``atomicity`` for deploying Custom Resource pods.
Atomicity in this context means that either all the Pods that are managed by a Custom Resource instance are deployed or none are. There is no intermediate state where some of the Custom Resource Pods are deployed and others are not. The atomicity property is only achievable if Kubernetes is able to provide ``guaranteed`` QoS to all the Custom Resource Pods.


### Segregate application instances on different worker nodes

Kubernetes provides mechanism of [Pod Node Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)
This mechanism enables identifying the nodes on which a Pod should run.
A set of labels are specified on the Pod Spec which are matched by the scheduler with the labels on the nodes when making scheduling decisions. 
If your Custom Resource Pods need to be subject to this scheduling constraint then you will need to define the Custom Resource Spec to allow input of 
such labels. The Custom Controller will need to be implemented to pass these labels to the underlying Custom Resource Pod Spec.
From a multi-tenancy perspective such labels can be used to segregate Pods of different application instances on different cluster nodes, which might be needed, for example, to provide differentiated quality-of-service for certain application instances.


### Co-location of application Pods on a worker node

Kubernetes provides a mechanism of [Pod Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which enables defining Pod scheduling rules based on labels of other Pods that are running on a node. 
Consider if your Custom Resource Pods need to be provided with such affinity rules. If so, provide an attribute in your Custom Resource
Spec definition where such rules can be specified. The Custom Controller will need to be implemented to pass these rules to the Custom Resource Pod Spec. 
From a multi-tenancy perspective these affinity rules can be helpful to co-locate all the Pods of an application instance on a particular worker node.


### Prevent cross-traffic between application instances

Kubernetes [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) provide
a mechanism to define firewall rules to control network traffic for Pods. In case you need to restrict traffic
to your Custom Resource's Pods then define a Spec attribute in your Custom Resource Spec definition to
specify labels to be applied to the Custom Resource's underlying Pods. 
Implement your Operator to apply this label to the underlying Pods that will be created by the Operator as part of handling a Custom Resource. Then it will be possible for users to specify NetworkPolicies for your Custom Resource Pods.
From a multi-tenancy perspective these labels will help ensure that there is no cross-traffic between Pods that belong to different application instances.
