# Kubernetes Operator Maturity Model Guidelines for multi-Operator Stacks

Below we present the guidelines related to consumability, configurability, security, robustness, debuggability and portability of Operators in multi-Operator stacks towards creating workflows consisting of Custom Resources.


## Consumability

Consumability guidelines focus on accessibility properties of Custom Resources 
that make it easy for application developers to build Custom Resource workflows.

[1. Design Custom Resource as a declarative API and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#design-custom-resource-as-a-declarative-api-and-avoid-inputs-as-imperative-actions)

[2. Make Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#make-custom-resource-type-definitions-compliant-with-kube-openapi)

[3. Consider to use kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#consider-to-use-kubectl-as-the-primary-interaction-mechanism)

[4. Define Man page for your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-man-page-for-your-custom-resources)


## Configurability

Configurability guidelines focus on configuration and customization of Custom Resource workflows. 

[5. Define inter Custom Resource binding information in Spec Properties](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-inter-custom-resource-binding-information-in-spec-properties)

[6. Document labels and annotations to be used with your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-labels-and-annotations-to-be-used-with-your-operator)

[7. Define Resource limits and Resource requests for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-resource-limits-and-resource-requests-for-custom-resources)

[8. Use ConfigMap or Custom Resource Annotation or Custom Resource Spec definition for underlying resource configuration](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#use-configmap-or-custom-resource-annotation-or-custom-resource-spec-definition-for-underlying-resource-configuration)


## Security

Security guidelines focus on Operator support of multi-tenancy and ability to define appropriate authorization controls
for workflows.

[9. Document Service Account needs of your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-service-account-needs-of-your-operator)

[10. Evaluate Service Account needs for Custom Resource Pods](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#evaluate-service-account-needs-for-custom-resource-pods)

[11. Define SecurityContext and PodSecurityPolicies for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-securitycontext-and-podsecuritypolicies-for-custom-resources)

[12. Make Custom Controllers Namespace aware](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#make-custom-controllers-namespace-aware)

[13. Define Custom Resource Node Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-node-affinity-rules)

[14. Define Custom Resource Pod Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-pod-affinity-rules)

[15. Define NetworkPolicy for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-networkpolicy-for-custom-resources)


## Robustness

Robustness guidelines focus on ensuring stability of workflows with respect to their
resource needs, their interdependencies, input validations, and garbage collection.

[16. Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[17. Document Resource dependency information of your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-resource-dependency-information-of-your-custom-resources)

[18. Include CRD installation hints in Helm chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#include-crd-installation-hints-in-helm-chart)

[19. Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition-yaml)

[20. Design for robustness against side-car injection into Custom Resource Pods](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#design-for-robustness-against-side-car-injection-into-custom-resource-pods)

[21. Define Custom Resource Anti-Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-anti-affinity-rules)

[22. Define Custom Resource Taint Toleration rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-taint-toleration-rules)

[23. Define PodDisruptionBudget for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-poddisruptionbudget-for-custom-resources)


## Debuggability

Debuggability guidelines focus on Operator and Custom Resource properties that enable easy debugging of Custom Resource workflows.

[24. Enable Audit logs for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#enable-audit-logs-for-custom-resources)

[25. Decide Custom Resource Metrics Collection strategy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#decide-custom-resource-metrics-collection-strategy)

[26. Expose Custom Resource Composition Information](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#expose-custom-resource-composition-information)

[27. Document how your Operator uses namespaces](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-how-your-operator-uses-namespaces)


## Portability

Portability guidelines focus on Operator and Custom Resource properties that enable deploying the Operators and workflows
on any Kubernetes distribution, on-prem or on cloud.

[28. Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#package-operator-as-helm-chart)

[29. Register CRDs as YAML Spec in Helm chart rather than in Operator code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#register-crds-as-yaml-spec-in-helm-chart-rather-than-in-operator-code)

[30. Use Kubernetes-native Certification Management Solution](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#use-kubernetes-native-certification-management-solution)


# Detail Guidelines


## Consumability

### Design Custom Resource as a declarative API and avoid inputs as imperative actions

A declarative API is one in which you specify the desired state of the software that your Operator is managing using the Custom Resource Spec definition. Prefer declarative specification over any imperative actions in Custom Resource Spec  definition. In multi-Operator setups this requirement ensures correct composability of workflows created from multiple Custom resources. Also, having all the Custom Resource Spec inputs as declarative values provides easy way to determine the state of the overall workflow.


### Make Custom Resource Type definitions compliant with Kube OpenAPI

Kubernetes API details are documented using Swagger v1.2 and OpenAPI. [Kube OpenAPI](https://github.com/kubernetes/kube-openapi) supports a subset of OpenAPI features to satisfy kubernetes use-cases. As Operators extend Kubernetes API, it is important to follow Kube OpenAPI features to provide consistent user experience.
Following actions are required to comply with Kube OpenAPI.

Add documentation on your Custom Resource Type definition and on the various fields in it.
The field names need to be defined using following pattern:
Kube OpenAPI name validation rules expect the field name in Go code and field name in JSON to be exactly 
same with just the first letter in different case (Go code requires CamelCase, JSON requires camelCase).

When defining the types corresponding to your Custom Resources, you should use kube-openapi annotation — ``+k8s:openapi-gen=true`` in the type definition to enable generating OpenAPI Spec documentation for your Custom Resources. An example of this annotation on type definition on CloudARK sample Postgres Custom Resource is shown below:
```
  // +k8s:openapi-gen=true
  type Postgres struct {
    :
  }
```

Defining this annotation on your type definition would enable Kubernetes API Server 
to generate documentation for your Custom Resource Spec properties. This can then be viewed using ``kubectl explain`` command. In multi-Operator setups, this guideline ensures consistency in how users can discover Spec properties of different Custom Resources needed in their workflows.


### Consider to use kubectl as the primary interaction mechanism

Custom Resources introduced by your Operator will naturally work with kubectl.
However, there might be operations that you want to support for which the declarative nature of Custom Resources
is not appropriate. An example of such an action is historical record of how Postgres Custom Resource has evolved over time
that might be supported by the Postgres Operator. Such an action does not fit naturally into the declarative
format of Custom Resource Definition. For such actions, we encourage you to consider using Kubernetes
extension mechanisms of Aggregated API servers and Custom Sub-resources. These mechanisms 
will allow you to continue using kubectl as the primary interaction point for your Operator.
Refer to [this blog post](https://medium.com/@cloudark/comparing-kubernetes-api-extension-mechanisms-of-custom-resource-definition-and-aggregated-api-64f4ca6d0966) to learn more about them. Before considering to introduce new CLI for your Operator, validate if you can use 
Kubernetes's built-in mechanisms instead. This is especially true in multi-Operator setups. Having to use different CLIs will not be a great user experience. 


### Define Man page for your Custom Resources

Application developers will need to know details about how to use Custom Resources of your Operator. 
This information goes beyond what is available through Custom Resource Spec properties.
Consider creating a Unix style ``man page`` for your Custom Resources.
For each Custom Resource include following information - if there are any implementation assumptions made by the Operator author, at a high-level how to use the Custom Resource, service account and RBAC needs of your Operator, service account and RBAC needs of your Custom Resource's Pods, any hard coded values such as those for resource requests/limits, disruption budgets, service accounts, tolerations, etc. In order to enable users to discover this information in Kubernetes-native manner, you can use our 'platform-as-code/usage' annotation on the CRD definition YAML. Create a ConfigMap with the usage information 
and add following annotation on your CRD definition YAML: 'platform-as-code/usage'. Set the value of this
annotation to the name of the ConfigMap that you have created with the usage information. Once this is done,
users will be able to access this information in Kubernetes-native manner using 'kubectl man <Custom Resource>'
command (once KubePlus API Add-on is installed by cluster administrator in your cluster).
If all the Operators in a multi-Operator stack support man pages for their Custom Resources, it will help application developers when creating their Custom Resource workflows.


## Configurability


### Define inter Custom Resource binding information in Spec Properties

Define Spec properties in Custom Resources that can be used to define the binding between them. This enables creating Kubernetes-native workflows by composing various Custom Resources together.


### Document labels and annotations to be used with your Operator

Your Operator may need certain labels or annotations to be added on Kubernetes built-in resources for its operation. Document this information as part of your Operator's man page. This information will be critical for correct operation
of Custom and built-in resource workflows. 


### Define Resource limits and Resource requests for Custom Resources

Kubernetes provides mechanism of [requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-types) for specifying the cpu and memory resource needs of a Pod's containers. When specified, Kubernetes scheduler ensures that the Pod is scheduled on a Node that has enough capacity 
for these resources. A Pod with request and limits specified for every container is given ``guaranteed`` Quality-of-Service (QoS) by the Kubernetes scheduler. A Pod in which only resource requests are specified for at least one container is given ``burstable`` QoS. A Pod with no requests/limits specified is given ``best effort`` QoS.
It is important to define resource requests and limits for your Custom Resources.
This helps with ensuring ``atomicity`` for deploying workflows of Custom Resources.
Atomicity in this context means that either all the Custom Resources in a workflow are deployed or none are.
There is no intermediate state where some of the Custom Resources are deployed and others are not.
The atomicity property is only achievable if Kubernetes is able to provide ``guaranteed`` QoS to all
the Custom Resources in a workflow. This is only possible if each Custom Resource defines resource requests and limits which need to be propagated to the underlying Pod's containers by the Operator.


### Use ConfigMap or Custom Resource Annotation or Custom Resource Spec definition for underlying resource configuration

An Operator generally needs to take configuration parameter as inputs for the underlying resource that it is managing.
We have seen three different approaches being used towards this in the community: using ConfigMaps, using Annotations, or using Spec definition itself. There are pros and cons of each of these approaches.
An advantage of using Spec properties is that it is possible to define validation rules for them.
This is not possible with annotations or ConfigMaps. An advantage of using ConfigMaps is that it supports
providing larger inputs (such as entire configuration files). This is not possible with Spec properties or annotations. An advantage of annotations is that adding/updating new fields does not require changes to the Spec.
Any of these approaches should be fine based on your Operator design. 
It is also possible that you may end up using multiple approaches, such as a ConfigMap with its name specified in the 
Custom Resource Spec definition.

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).

In any case, choosing one of these three mechanisms is better than other approaches, such as using a initContainer to download config file/data from remote storage (s3, github, etc.), or using Pod environment variables. The problem with these approaches is that at the workflow level they can lead to inconsistency of providing inputs to different Custom Resources. This in turn can make it difficult to redeploy/reproduce the workflows on different clusters.


## Security

### Document Service Account needs of your Operator

Your Operator may be need to use a specific service account with specific permissions. Clearly document the service account needs of your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD. In multi-Operator stacks, knowing the service accounts and their RBAC
permissions enables users to know the security posture of the stack.
Be explicit in defining only the required permissions and nothing more. This ensures that the cluster is safe against
unintended actions by any of the Operators (malicious/byzantine actions of compromised Operators 
or benign faults appearing in the Operators).


### Evaluate Service Account needs for Custom Resource Pods

Your Custom Resource's Pods may need to run with specific Service account. If that is the case, one
of the decisions you will need to make is whether that Service account should be provided by application
developers. If so, provide an attribute in Custom Resource Spec
definition to define a Service account. Alternatively, if the Custom Controller is hard coding
the Service account in the Pod Spec, then surface this information through the Custom Resource ``man page``.
If all the Custom Resources have service accounts defined, it enables users to understand the complete security
posture of the workflows.


### Define SecurityContext and PodSecurityPolicies for Custom Resources

Kubernetes provides mechanism of [SecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) that can be used to define the security attributes of a Pod's containers (userID, groupID, Linux capabilities, etc.). In your Operator implementation, you may decide to create Custom Resource Pods using
certain settings for the securityContext. Surface these settings through the Custom Resource ``man page``. The value of this annotation should be the name of a ConfigMap that contains the security context attributes defined as the data values. Make sure to include this ConfigMap in your Operator's Helm chart.
By surfacing the security context information in this way, it will be possible for the DevOps engineers and
Application developers to find out the security context with which Custom Resource Pods will run in the cluster.

Kubernetes [Pod Security Policies](https://kubernetes.io/docs/concepts/policy/pod-security-policy/) provide a way to define policies related to creating privileged Pods in a cluster.
As part of handling Custom Resources, if your Operator is creating privileged Pods then make sure
to surface this information as part of the Custom Resource man page. If you want
to provide the control of creating privileged Pods to Custom Resource user then define a Spec attribute
for this purpose in your Custom Resource Spec definition.

If every Operator and Custom Resource is implemented to take SecurityContext and PodSecurityPolicies (PSPs) as input
through Spec definition, then it will be possible to define a uniform security context and PSP for the entire
workflow. If on the other hand, some of the Operators have hard-coded these but they surface it through
the Operator's man page then it will be possible for application developers to understand the overall security posture of the entire workflow.


### Make Custom Controllers namespace aware

Your Operator should support creating resources within different namespaces rather than just in the namespace
in which it is deployed (or the default namespace). This will allow your Operator to support 
multi-tenancy through namespaces. When all the Custom Resources and Operators support namespaces, it is possible
to create multi-tenant workflows.


### Define Custom Resource Node Affinity rules

Kubernetes provides mechanism of [Pod Node Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)
This mechanism enables identifying the nodes on which a Pod should run.
A set of labels are specified on the Pod Spec which are matched by the scheduler 
with the labels on the nodes when making scheduling decision. 
If your Custom Resource Pods need to be
subject to this scheduling constraint then you will need to define the Custom Resource Spec to allow input of 
such labels. The Custom Controller will need to be implemented to pass these labels to the Custom Resource Pod Spec.
In multi-Operator stacks if every Custom Resource supports node affinity rules, it will enable co-locating entire workflows on specific nodes.


### Define Custom Resource Pod Affinity rules

Kubernetes provides mechanism of [Pod Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which enables defining Pod scheduling rules based on labels of other Pods that are running on a node. 
Consider if your Custom Resource Pods need to be provided with such affinity rules corresponding
to other Custom Resources from same or other Operator. If so, provide an attribute in your Custom Resource
Spec definition where such rules can be specified. The Custom Controller will need to be implemented to pass these 
rules to the Custom Resource Pod Spec. In multi-Operator stacks if every Custom Resource supports pod affinity rules, it will enable co-locating entire workflows together on a node.


### Define NetworkPolicy for Custom Resources

Kubernetes [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) provide
a mechanism to define firewall rules to control network traffic for Pods. In case you need to restrict traffic
to your Custom Resource's Pods then define a Spec attribute in your Custom Resource Spec definition to
specify labels to be applied to the Custom Resource's underlying Pods. 
Implement your Operator to apply this label to the underlying Pods that will be created by the Operator as part of handling a Custom Resource. Then it will be possible for users to specify NetworkPolicies for your Custom Resource Pods.
If you don't want to modify Custom Resource Spec properties, use KubePlus API-on which supports ``Fn::AddLabel``function which can be used to add labels to Custom Resource Pods. In multi-Operator stacks if every Custom Resource supports NetworkPolicy labels then it will enable enforcing traffic restriction for entire workflow. 

## Robustness

### Set OwnerReferences for underlying resources owned by your Custom Resource

An Operator will typically create one or more native Kubernetes resources as part of instantiating a Custom Resource instance. Set the OwnerReference attribute of such underlying resources to the Custom Resource instance that is
being created. OwnerReference setting is essential for correct garbage collection of Custom Resources. 
It also help with finding runtime composition tree of your Custom Resource instances.
If all the Operators are correctly handling OwnerReferences, then the garbage collection benefit will get
extended for the entire workflow consisting of Custom Resources.


### Document Resource dependency information of your Custom Resources

In order to use your Custom Resource correctly, other Kubernetes built-in resource or Custom Resource from same
or different Operator may need to be created first. Explicitly document such dependencies as part of your Operator's man page. This information will be useful for Application developers in creating workflows consisting of Custom and built-in
resources. You can also use our KubePlus API Add-on which provides ``PlatformStack`` CRD to capture this type of dependency information. KubePlus API Add-on uses this information for several purposes such as - preventing out-of-order creation of Kubernetes resources, ensuring that Custom or built-in resources that are in use as part of some workflows are not inadvertently deleted. 


### Include CRD installation hints in Helm chart 

Helm 2.0 defines crd-install hook that directs Helm to install CRDs first before installing rest of your
Helm chart that might refer to the Custom Resources defined by the CRDs. 

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: moodles.moodlecontroller.kubeplus
    annotations:
      helm.sh/hook: crd-install
```

In Helm 3.0 the crd-install annotation is no longer supported. Instead, a separate directory
named ``crd`` needs to be created as part of the Helm chart directory in which all the CRDs
are to be specified. By defining CRDs inside this directory, Helm 3.0 guarantees to install
the CRDs before installing other templates, which may consist of Custom Resources introduced by that CRD.
Installing CRDs first is important -- otherwise creation of Custom Resource workflows will fail as the Kubernetes cluster control plane won't be able to recognize the Custom Resources used in your workflows.


### Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML

Your Custom Resource Spec definitions will contain different properties and they may have some
domain-specific validation requirements. Kubernetes 1.13 onwards you will be able to use 
OpenAPI v3 schema to define validation requirements for your Custom Resource Spec. For instance,
below is an example of adding validation rules for our sample Postgres CRD. The rules define that
the Postgres Custom Resource Spec properties of 'databases' and 'users' should be of type Array
and that every element of this array should be of type String. Once such validation rules are defined,
Kubernetes will reject any Custom Resource Specs that do not satisfy these requirements.

```
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: postgreses.postgrescontroller.kubeplus
  annotations:
    helm.sh/hook: crd-install
    platform-as-code/composition: Deployment, Service
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

In the context of Custom Resource workflows, this requirement ensures that the entire workflow specifications
can be correctly validated.  

### Design for robustness against side-car injection into Custom Resource Pods

Certain Operators such as those that take Volume backups work by injecting side-car containers into the
Custom Resource's Pods for which the Volume backup is being requested. Such an operation leads to restarting
of those Pods, which can lead to intermittent failure of the workflow in which that Custom Resource is being used.
To prevent against such situations, design your Operator to subscribe to Pod restart events for your Custom Resources
and ensure that the required Custom Resource Spec properties are maintained on the Pod.


### Define Custom Resource Anti-Affinity rules

Kubernetes provides mechanism of [Pod Anti-Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which are opposite of Pod affinity rules. 
Consider if your Custom Resource Pods need to be provided with such anti-affinity rules corresponding to 
other Custom Resoures from other Operators. If so, provide an attribute in your Custom Resource
Spec definition where such rules can be specified. Implement the Custom Controller to pass these 
rules to the Custom Resource Pod Spec. These rules will be useful in deploying different Custom Resource workflows on separate nodes.


### Define Custom Resource Taint Toleration rules

Kubernetes provides mechanism of [taints and tolerations](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) to restrict scheduling of Pods on certain nodes. If you want your Custom Resource pods
to be able to tolerate the taints on a node, define an attribute in your Custom Resource Spec definition
where such tolerations can be specified. Implement the Custom Controller to pass the toleration labels to
the Custom Resource Pod Spec. These rules will be useful in deploying certain Custom Resource workflows 
on dedicated nodes which might be needed, for example, to provide differentiated quality-of-service for certain workflows.


### Define PodDisruptionBudget for Custom Resources

Kubernetes provides mechanism of [Pod Disruption Budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) (PDB) that can be used to define the disruption tolerance for Pods. Specifically, two
fields are provided - 'minAvailable' and 'maxUnavailable'. minAvailable is the minimum number of Pods that 
should be always running in a cluster. maxUnavailable is complementary and defines the maximum number of Pods
that can be unavailable in a cluster. These two fields provide a way to control the availability of Pods in a cluster.
When implementing the controller for your Custom Resource, carefully consider such availability requirements for your Custom Resource Pods. If you decide to implement PDB for your Custom Resource Pods, consider whether
the disruption budget should be set by application developers. 
If yes, then ensure that the Custom Resource Spec definition has a field to specify a disruption budget.
Implement the Custom Controller to pass this value to the Pod Spec.
If on the other hand you decide to hard code this choice in your Custom Controller implementation then
surface it to Custom Resource users throught its ``man page``.
In multi-Operator stacks, if each Custom Resource contains PDB in its Spec it provides guarantee of stability
of Custom Resource workflows as a whole. Without PDBs, it may happen that one Custom Resource gets disrupted if the Kubernetes scheduler comes under resource pressure causing entire workflow to fail.


## Debuggability

### Enable Audit logs for Custom Resources

Kubernetes API server allows creation of audit logs for incoming requests. This guideline is intended for cluster
administrators. It requires that cluster administrators define the Audit Policy to include tracking of all the Custom 
Resources available in a cluster. Using the ``RequestResponse`` audit level is useful as that enables keeping track
of the incoming Custom Resource requests and the responses sent by the API server.
Once audit logs are enabled, it will be possible to use tools like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) to query history of Custom Resources and thus track the evolution of Custom Resource workflows over time.


### Decide Custom Resource Metrics Collection strategy

Plan for metrics collection of Custom Resources managed by your Operator. This information is useful for understanding effect of various actions on your Custom Resources over time for their traceability. 
For example, [this MySQL Operator](https://github.com/oracle/mysql-operator/) 
collects information such as how many clusters were created. One option to collect metrics 
is to build the metrics collection inside your Custom Controller itself, as done by the MySQL Operator.
Another option is to leverage Kubernetes Audit Logs for this purpose. 
Then, you can use external tooling like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) 
to build the required metrics. Once metrics are collected, you should consider exposing them in Prometheus format.


### Expose Custom Resource Composition Information

When using Custom Resources of your Operator, Application developers will often need help with debugging
Custom Resources when failures occur. One of the key things as part of this is to know what Kubernetes's
native resources are created by the Operator when instantiating a Custom Resource instance.
It is helpful to document this information as part of your Operator's documentation. 
In order to enable users to discover this information in Kubernetes-native manner, you can use
our 'platform-as-code/composition' annotation on the CRD definition YAML. The value of this annotation
is the list of all the Kubernetes resources such as Deployment, ConfigMap, Secret, etc. that an Operator
creates as part of creating a Custom Resource instance. Once this is done, users will be able to use
the 'kubectl composition <Custom Resource> <Custom Resource instance>' command to find the runtime
composition tree of all the Kubernetes resource that are created by the Operator as part of instantiating
a Custom Resource instance. This information can then aid debugging of the workflows.
Here is an example of defining 'usage' and 'composition' annotations on Moodle CRD.

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: moodles.moodlecontroller.kubeplus
    annotations:
      helm.sh/hook: crd-install
      platform-as-code/usage: moodle-operator-usage.usage
      platform-as-code/composition: Deployment, Service, PersistentVolume, PersistentVolumeClaim, Secret, Ingress
```

By defining composition annotation for every CRD, it will be possible to find workflow-level resource composition tree. This will aid in visualization of runtime structure of the workflows.


### Document how your Operator uses namespaces

For Operator developers it is critical to consider how their Operator works with namespaces. Typically, an Operator can be installed in one of the following configurations:

  * Operator runs in the default namespace and Custom Resource instances are created in the default namespace.

  * Operator runs in the default namespace but Custom Resource instances can be created in non-default namespaces.

  * Operator runs in a non-default namespace and Custom Resource instances can be created in that namespace.

Given these options, it will help consumers of your Operator if there is a clear documentation of how namespaces are used by your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD so that this information will be available as part of the Custom Resource man page.
If this information is defined for every Operator installed in the cluster, it will be possible for application developers to understand support for namespaces for Custom Resource workflows as a whole.


## Portability

### Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * All Custom Resource Definitions for Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * ConfigMaps corresponding to Platform-as-Code annotations that you have added on your Custom Resource Definition (CRD).

Use Helm chart's Values.yaml for Operator customization. 
By defining Helm charts, multi-Operator stacks and Custom Resource workflows can be created on any cloud or Kubernetes distribution.


### Register CRDs as YAML Spec in Helm chart rather than in Operator code

Installing CRD requires Cluster-scope permission. If the CRD registration is done as YAML manifest, then it is possible to separate CRD registration from the Operator Pod deployment. CRD registration
can be done by cluster administrator while Operator Pod deployment can be done by a non-admin user. 
It is then possible to deploy the Operator in different namespaces with different customizations.
On the other hand, if CRD registration is done as part of your Operator code then the deployment of the Operator Pod will need Cluster-scope permissions. This will tie together installation of CRD with that of the Operator, which
may not be the best setup in certain situations. Another reason to register CRD as YAML is because kube-openapi validation can be defined as part of it.


### Use Kubernetes-native Certification Management Solution

As part of creating Custom Resource workflows, there might be a need to install SSL certificates for the workflows.
Using a Kubernetes-native solution in the form of an SSL Operator will make the workflows portable across cloud providers
and Kubernetes distributions.




