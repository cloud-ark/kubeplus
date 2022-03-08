# Kubernetes Operator Maturity Model:  Operator readiness for multi-tenant environments

Kubernetes Operators extend Kubernetes Resource set (APIs) for managing softwares like databases, key-value stores, API gateways, etc. in Kubernetes native manner. Today enterprise DevOps teams are using multiple Kubernetes Operators on their clusters to create custom PaaSes. A PaaS requires an Operator’s readiness for multi-tenant environments. Towards this we have developed following set of Operator Maturity Model guidelines. The model is divided into three categories with a set of guidelines in each category. 


## Summary of the Guidelines

## Consumability

Consumability guidelines focus on design of Custom Resources.

[1. Design Custom Resource as a declarative API and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#design-custom-resource-as-a-declarative-api-and-avoid-inputs-as-imperative-actions)

[2. Make Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#make-custom-resource-type-definitions-compliant-with-kube-openapi)

[3. Consider using kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#consider-using-kubectl-as-the-primary-interaction-mechanism)

[4. Expose Operator developer’s assumptions and requirements about Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#expose-operator-developers-assumptions-and-requirements-about-custom-resources)

[5. Provide a way to customize underlying software](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#provide-a-way-to-customize-underlying-software)

[6. Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#package-operator-as-helm-chart)

[7. Register CRDs as YAML Spec in Helm chart rather than in Operator code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#register-crds-as-yaml-spec-in-helm-chart-rather-than-in-operator-code)

[8. Include CRD installation hints in Helm chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#include-crd-installation-hints-in-helm-chart)


## Security

Security guidelines help in identifying authorization controls required in building multi-tenant application stacks using Kubernetes resources (built-in and custom).

[9. Define Service Account for Operator Pod](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-service-account-for-operator-pod)

[10. Define Service Account for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-service-account-for-custom-resources)

[11. Define SecurityContext and PodSecurityPolicies for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-securitycontext-and-podsecuritypolicies-for-custom-resources)

[12. Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[13. Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition-yaml)



## SaaS multi-tenancy

[14. Namespace based multi-tenancy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#namespace-based-multi-tenancy)

[15. Atomicity of application instance deployment](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#atomicity-of-application-instance-deployment)

[16. Segregate application instances on different worker nodes](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#segregate-application-instances-on-different-worker-nodes)

[17. Co-location of application Pods on a worker node](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#co-location-of-application-pods-on-a-worker-node)

[18. Prevent cross-traffic between application instances](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#prevent-cross-traffic-between-application-instances)


# Detail Guidelines


## Consumability

Consumability guidelines focus on design of Custom Resources (APIs) to simplify their use in building application stacks.

### Design Custom Resource as a declarative API and avoid inputs as imperative actions

A declarative API is one in which you specify the desired state of the software that your Operator is managing using the Custom Resource Spec definition. Prefer declarative specification over any imperative actions in Custom Resource Spec definition as declarative values provides easy way to determine the current 
state of the application.

### Make Custom Resource Type definitions compliant with Kube OpenAPI

Kubernetes API details are documented using Swagger v1.2 and OpenAPI. [Kube OpenAPI](https://github.com/kubernetes/kube-openapi) supports a subset of OpenAPI features to satisfy kubernetes use-cases. As Operators extend Kubernetes API, it is important to follow Kube OpenAPI features to provide consistent user experience.
Following actions are required to comply with Kube OpenAPI.

- Add documentation on your Custom Resource Type definition and on the various fields in it. The field names need to be defined using following pattern:
Kube OpenAPI name validation rules expect the field name in Go code and field name in JSON to be exactly same with just the first letter in different case (Go code requires CamelCase, JSON requires camelCase).
- When defining the types corresponding to your Custom Resources, you should use kube-openapi annotation — ``+k8s:openapi-gen=true`` in the type definition to enable generating OpenAPI Spec documentation for your Custom Resources. An example of this annotation on type definition on CloudARK sample Postgres Custom Resource is shown below:
```
  // +k8s:openapi-gen=true
  type Postgres struct {
    :
  }
```

Defining this annotation on your type definition would enable Kubernetes API Server to generate documentation for your Custom Resource Spec properties. This can then be viewed using ``kubectl explain`` command.

### Consider using kubectl as the primary interaction mechanism

Before introducing new CLI for your Operator, validate if you can use Kubernetes's built-in mechanisms instead. This is especially true in multi-Operator setups. Having to use different CLIs will not be a great user experience.

Custom Resources introduced by your Operator will naturally work with kubectl.
However, there might be operations that you want to support for which the declarative nature of Custom Resources is not appropriate. For such operations see if you can develop [``kubectl plugins``](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).


### Expose Operator developer’s assumptions and requirements about Custom Resources

Application developers will need to know details about how to use Custom Resources of your Operator. This information goes beyond what is available through Custom Resource Spec properties. Here are some of the Operator developer assumptions that are not captured in your Custom Resource spec definition:
- what sub-resources will be created by the Operator as part of handling the Custom Resource.
- your Custom Resources may need certain labels or annotations to be added on other Kubernetes resources for the Operator's functioning.
- your Custom Resource may also depend on spec property of another Kubernetes resource.
- service account and RBAC needs of your Operator, service account and RBAC needs of your Custom Resource's Pods.
- any hard-coded values such as those for resource requests/limits, Pod disruption budgets, taints and tolerations, etc. that are defined for your Custom Resource.

We have devised an easy way to capture these assumptions as CRD annotations. We currently offer following CRD annotations for this purpose:
```
resource/usage
resource/composition
resource/annotation-relationship
resource/label-relationship
resource/specproperty-relationship
```
Once these annotations are defined on your Operator's CRDs, application developers can find out fine-grained information about your Custom Resources by using 
[our KubePlus kubectl plugins](https://github.com/cloud-ark/kubeplus#2-kubeplus-kubectl-plugins). [Here](https://github.com/cloud-ark/kubeplus/blob/master/Operator-annotations.md) you can find annotations that we maintain for community Operators.


### Provide a way to customize underlying software

An Operator generally needs to take configuration parameter as inputs for the underlying resource that it is managing. We have seen three different approaches being used towards this in the community: using ConfigMaps, using Annotations, or using Custom Resource Spec definition itself. There are pros and cons of each of these approaches. An advantage of using Spec properties is that it is possible to define validation rules for them. This is not possible with annotations or ConfigMaps. An advantage of using ConfigMaps is that it supports providing larger inputs (such as entire configuration files). This is not possible with Spec properties or annotations. An advantage of annotations is that adding/updating new fields does not require changes to the Spec. Any of these approaches should be fine based on your Operator design. It is also possible that you may end up using multiple approaches, such as a ConfigMap with its name specified in the Custom Resource Spec definition.

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22). DataStax's [Cassandra Operator](https://github.com/datastax/cass-operator) uses Spec definition.

Choosing one of these three mechanisms is better than other approaches, such as using a initContainer to download config file/data from remote storage (s3, github, etc.), or using Pod environment variables for passing in configuration data for the underlying resource. The problem with these approaches is that at the application stack level they can lead to inconsistency of providing inputs to different Custom Resources. This in turn can make it difficult to redeploy/reproduce the stacks on different clusters.

### Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * All Custom Resource Definitions for Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * ConfigMaps corresponding to 'resource/usage' annotation that you have added on your Custom Resource Definition (CRD).

Use Helm chart's Values.yaml for Operator customization.


### Register CRDs as YAML Spec in Helm chart rather than in Operator code

Installing CRD requires Cluster-scope permission. If the CRD registration is done as YAML manifest, then it is possible to separate CRD registration from the Operator Pod deployment. CRD registration can be done by cluster administrator while Operator Pod deployment can be done by a non-admin user. It is then possible to deploy the Operator in different namespaces with different customizations.
On the other hand, if CRD registration is done as part of your Operator code then the deployment of the Operator Pod will need Cluster-scope permissions. This will tie together installation of CRD with that of the Operator, which
may not be the best setup in certain situations. Another reason to register CRD as YAML is because kube-openapi validation can be defined as part of it.


### Include CRD installation hints in Helm chart 

Installing CRDs first is important -- otherwise deploying application manifests consisting of Custom Resources will fail as the Kubernetes cluster control plane won't be able to recognize the Custom Resources used in your application.

Helm 2.0 defines crd-install hook that directs Helm to install CRDs first before installing rest of your Helm chart that might refer to the Custom Resources defined by the CRDs. 

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: moodles.moodlecontroller.kubeplus
    annotations:
      helm.sh/hook: crd-install
```

In Helm 3.0 the crd-install annotation is no longer supported. Instead, a separate directory named ``crds`` needs to be created as part of the Helm chart directory in which all the CRDs are to be specified. By defining CRDs inside this directory, Helm 3.0 guarantees to install the CRDs before installing other templates, which may consist of Custom Resources introduced by that CRD.




## Security

Security guidelines help in applying authorization controls in building multi-tenant application stacks.

### Define Service Account for Operator Pod

Your Operator may be need to use a specific service account with specific permissions. Clearly document the service account needs of your Operator Pod. Include this information in the ConfigMap that you will add for the 'resource/usage' annotation on the CRD. In multi-Operator setups, knowing the service accounts and their RBAC permissions enables users to know the security posture of their environment. Be explicit in defining only the required permissions and nothing more. This ensures that the cluster is safe against unintended actions by any of the Operators (malicious/byzantine actions of compromised Operators or benign faults appearing in the Operators).


### Define Service Account for Custom Resources

Your Custom Resource's Pods may need to run with specific service account permissions. If that is the case, one of the decisions you will need to make is whether that service account should be provided by DevOps engineers. If so, provide an attribute in Custom Resource Spec definition to define the service account. Alternatively, if the Custom Controller is hard coding the service account name in the underlying Pod's Spec, then expose this information as Operator’s assumption using ``resource/usage`` CRD annotation discussed above. If all the Custom Resources have service accounts defined, it enables users to understand the complete security posture of their application stack.


### Define SecurityContext and PodSecurityPolicies for Custom Resources

Kubernetes provides mechanism of [SecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) that can be used to define the security attributes of a Pod's containers (userID, groupID, Linux capabilities, etc.). In your Operator implementation, you may decide to create Custom Resource Pods using
certain settings for the securityContext. Surface these settings through the Custom Resource the ``resource/usage`` CRD annotation. By surfacing the security context information in this way, it will be possible for the DevOps engineers and
Application developers to find out the security context with which Custom Resource Pods will run in the cluster.

Kubernetes [Pod Security Policies](https://kubernetes.io/docs/concepts/policy/pod-security-policy/) provide a way to define policies related to creating privileged Pods in a cluster.
As part of handling Custom Resources, if your Operator is creating privileged Pods then make sure
to surface this information through ``resource/usage`` CRD annotation. If you want
to provide the control of creating privileged Pods to Custom Resource user then define a Spec attribute for this purpose in your Custom Resource Spec definition.


### Set OwnerReferences for underlying resources owned by your Custom Resource

An Operator will typically create one or more native Kubernetes resources as part of instantiating a Custom Resource instance. Set the OwnerReference attribute of such underlying resources to the Custom Resource instance that is
being created. OwnerReference setting is essential for correct garbage collection of Custom Resources. 
It also helps with finding runtime composition tree of your Custom Resource instances.
If all the Operators are correctly handling OwnerReferences, then the garbage collection benefit will get
extended for the entire application stack consisting of Custom Resources.


### Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML

Your Custom Resource Spec definitions will contain different properties and they may have some domain-specific validation requirements. Kubernetes 1.13 onwards you will be able to use OpenAPI v3 schema to define validation requirements for your Custom Resource Spec. For instance, below is an example of adding validation rules for our sample Postgres CRD. The rules define that the Postgres Custom Resource Spec properties of 'databases' and 'users' should be of type Array and that every element of this array should be of type String. Once such validation rules are defined, Kubernetes will reject any Custom Resource Specs that do not satisfy these requirements.

```
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: postgreses.postgrescontroller.kubeplus
  annotations:
    helm.sh/hook: crd-install
    resource/composition: Deployment, Service
spec:
  group: postgrescontroller.kubeplus
  version: v1
  names:
    kind: Postgres
    plural: postgreses
  scope: Namespaced
validation:
   # openAPIV3Schema is the schema for validating custom objects.
    openAPIV3Schema:
      properties:
        spec:
          properties:
            databases:
              type: array
              items:
                type: string
            users:
              type: array
              items:
                type: string 
```

In the context of application stacks, this requirement ensures that the entire  stack can be validated for correctness.



## SaaS multi-tenancy


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
