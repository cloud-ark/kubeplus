# Kubernetes Operator Guidelines to enable Multi-Operator Platform Stacks (Platform-as-Code)

Kubernetes Operators enable running third-party softwares directly on Kubernetes using Kubernetes
Custom Resources. Various Operators are being built today for variety of softwares such as 
MySQL, Postgres, Cassandra, Airflow, Redis, MongoDB, Kafka, Prometheus, Logstash, Moodle, Wordpress, Odoo, etc.
Different Operators are used to create custom platform stacks on a Kubernetes cluster. The Custom Resources
from the different Operators enable creating application 
[platform workflows as Code](https://cloudark.io/platform-as-code). Such codified platform workflows are Kubernetes-native; they don't have any platform vendor lock-in; and they can be created/re-created on any Kubernetes cluster.

We present below guidelines that ensure Operators and their Custom Resources are developed in way
such that they can be used seamlessly in such multi-Operator settings. 
Operators that are developed following these guidelines provide ease of installation, management, and discovery to 
users (DevOps Engineers and Application Developers).

Check out [this post](https://medium.com/@cloudark/analysis-of-open-source-kubernetes-operators-f6be898f2340) about our analysis of more than 100 open source Operators for their conformance to some of these guidelines.

## Introduction

The term Kubernetes Operator refers to a combination of one or more Kubernetes Custom Resources and their associated Custom Controllers. A Custom Resource provides declarative model to specify the desired state of a resource
that the associated Custom Controller is built to reconcile. For example, in order to run Cassandra clusters on top of
a Kubernetes cluster, one can define a Custom Resource to represent Cassandra cluster and the corresponding
Custom Controller to create and manage such clusters. Such a combination of Cassandra Custom Resource
and the Cassandra Custom controller can be thought of as an example of a Kubernetes Cassandra Operator.

The guidelines below are divided into following categories - Custom Resources, Custom Controllers, Operator packaging,
Operator and underlying resource configuration, Interactions and Discoverability, Documentation.

## Custom Resources

[1) Design Custom Resource as a declarative API and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#1-design-operator-with-declarative-apis-and-avoid-inputs-as-imperative-actions)

[2) Make Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#2-make-custom-resource-type-definitions-compliant-with-kube-openapi)

[3) Define Resource limits and Resource requests for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#3-define-resource-limits-and-resource-requests-for-custom-resources)

[4) Define PodDisruptionBudget for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#4-define-poddisruptionbudget-for-custom-resources)

[5) Define Custom Resource Node Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#5-define-custom-resource-node-affinity-rules)

[6) Define Custom Resource Anti-Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#6-define-custom-resource-anti-affinity-rules)

[7) Define Custom Resource Taint Toleration rules(https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#7-define-custom-resource-taint-toleration-rules)

[8) Define SecurityContext for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#8-define-securitycontext-for-custom-resource-pods)

[9) Evaluate Service Account needs for Custom Resource Pods](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#9-evaluate-service-account-needs-for-custom-resource-pods)


## Custom Controllers

[10) Make Custom Controllers Namespace aware](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#10-make-custom-controllers-namespace-aware)

[11) Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#11-set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[12) Decide Custom Resource Metrics Collection strategy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#12-decide-custom-resource-metrics-collection-strategy)


## Operator packaging

[13) Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#13-package-operator-as-helm-chart)

[14) Register CRDs as YAML Spec in Helm chart rather than in Operator code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#14-register-crds-as-yaml-spec-in-helm-chart-rather-than-in-operator-code)

[15) Add crd-install Helm hook annotation on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#15-add-crd-install-helm-hook-annotation-on-your-crd-yaml)

[16) Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#16-define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition-yaml)


## Operator and underlying resource configuration

[17) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#17-use-configmap-or-annotation-or-spec-definition-for-custom-resource-configurables)

[18) Use Helm chart or ConfigMap for Operator configurables (Helm chart related)](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#18-use-helm-chart-or-configmap-for-operator-configurables)


## Interaction and Discoverability

[19) Consider to use kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#19-consider-to-use-kubectl-as-the-primary-interaction-mechanism)

[20) Add Platform-as-Code annotations on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#20-add-platform-as-code-annotations-on-your-crd-yaml)


## Documentation

[21) Document how your Operator uses namespaces](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#21-document-how-your-operator-uses-namespaces)

[22) Document Service Account needs of your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#22-document-service-account-needs-of-your-operator)

[23) Document naming convention and labels to be used with your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#23-document-naming-convention-and-labels-to-be-used-with-your-custom-resources)


# Custom Resources

## 1) Design Operator with declarative API/s and avoid inputs as imperative actions

A declarative API is one in which you specify the desired state of your custom resource using the Custom Resource Spec definition. Prefer declarative specification over any imperative actions in Custom Resource Spec Type definition. Custom controller code should be written such that it reconciles the current state of the underlying software with the desired state by performing diff of the current state with the desired state. For example, when writing a Postgres Operator, the custom controller should be written to perform diff of the existing value of ‘users’ with the desired value of ‘users’ based on the received desired state and perform the required actions (such as adding new users, deleting current users, etc.). An example where underlying imperative actions are exposed in the Spec is this 
[MySQL Backup Custom Resource Spec](https://github.com/oracle/mysql-operator/blob/master/examples/backup/backup.yaml#L7). Here the fact that MySQL Backup is done using mysqldump tool is exposed in the Spec.
In our view such internal details should not be exposed in the Spec as it prevents Custom Resource Type definition 
to evolve independently without affecting its users.

## 2) Make Custom Resource Type definitions compliant with Kube OpenAPI

Kubernetes API details are documented using Swagger v1.2 and OpenAPI. [Kube OpenAPI](https://github.com/kubernetes/kube-openapi) supports a subset of OpenAPI features to satisfy kubernetes use-cases. As Operators extend Kubernetes API, it is important to follow Kube OpenAPI features to provide consistent user experience.
Following actions are required to comply with Kube OpenAPI.

Add documentation on your custom resource Type definition and on the various fields in it.
The field names need to be defined using following pattern:
Kube OpenAPI name validation rules expect the field name in Go code and field name in JSON to be exactly 
same with just the first letter in different case (Go code requires CamelCase, JSON requires camelCase).

When defining the types corresponding to your custom resources, you should use kube-openapi annotation — “+k8s:openapi-gen=true’’ in the type definition to enable generating OpenAPI Spec documentation for your custom resources. An example of this annotation on type definition on CloudARK sample Postgres custom resource is as follows:
```
  // +k8s:openapi-gen=true
  type Postgres struct {
    :
  }
```

## 3) Define Resource limit and Resource requests for Custom Resources

Kubernetes provides mechanism of [requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#resource-types) for specifying the cpu and memory resource needs of a Pod's containers. When specified, Kubernetes scheduler ensures that the Pod is scheduled on a Node that has enough capacity 
for these resources. A Pod with request and limits specified for every container is given ``guaranteed`` Quality-of-Service (QoS) by the Kubernetes scheduler. A Pod in which only resource requests are specified for at least one container is given ``burstable`` QoS. A Pod with no requests/limits specified is given ``best effort`` QoS.
If you do decide to implement requests and limits for the Custom Resource Pods, 
then the main decision you will need to take is whether to
allow Custom Resource users to provide this information. If you allow them to provide this information, 
then you will need to design the Custom Resource Spec to take the resource requests/limits as inputs. 
Your custom controller will need to be implemented to pass this information through to the Pod creation Spec. 
In case you decide to hard code this information in your
custom controller, then it will be useful to your Operator users to know about this implementation choice.
One way to surface this information is to include it in the definition of the 
Custom Resource YAML (CRD YAML) (check guideline #20 for details).


## 4) Define PodDisruptionBudget for Custom Resources

Kubernetes provides mechanism of [Pod Disruption Budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) (PDB) that can be used to define the disruption tolerance for Pods. Specifically, two
fields are provided - 'minAvailable' and 'maxUnavailable'. minAvailable is the minimum number of Pods that 
should be always running in a cluster. maxUnavailable is complementary and defines the maximum number of Pods
that can be unavailable in a cluster. These two fields provide a way to control the availability of Pods in a cluster.
They ensure that minimum number of Pods will always be present. When implementing the controller for your Custom Resource, carefully consider such availability requirements for your Custom Resource instance's Pods. If you do decide to implement PDB for your Custom Resource Pods, you will have to decide whether you want to expose this control to your Custom Resource 
users. If yes, then ensure that the Custom Resource Spec definition has a field to specify a disruption budget.
If on the other hand you decide to hard code this choice in your custom controller implementation then
surface it to your Operator users through Custom Resource Definition YAML (CRD YAML)
(check guideline #20 for details).


## 5) Define Custom Resource Node Affinity rules

Kubernetes provides mechanism of [Pod Node Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)
This mechanism enables defining the specific nodes on which a Pod should run.
The way this is achieved is by providing a set of labels on the Pod Spec that are matched by the scheduler 
with the labels on the nodes when making the scheduling decision. If your Custom Resource Pods need to be
subject to this scheduling constraint and if you want to give the control of specifying the labels to 
your users then you will need to define the Custom Resource Spec to allow input of such labels.


## 6) Define Custom Resource Anti-Affinity rules

Kubernetes also provides mechanism of [Pod Anti-Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which enables defining Pod scheduling rules
based on labels of other Pods that are running on a node. Consider if your Custom Resource Pods
need to be provided with such anti-affinity rules. If so, provide an attribute in the Custom Resource
Spec definition where such rules can be specified. 


## 7) Define Custom Resource Taint Toleration rules

Kubernetes provides mechanism of [taints and tolerations](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) to restrict scheduling of Pods on certain nodes. If you want your Custom Resource pods
to be able to tolerate the taints on a node, then define an attribute in your Custom Resource Spec definition
where such tolerations can be specified. 


## 8) Define SecurityContext for Custom Resources

Kubernetes provides mechanism of [SecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) that can be used to define the security attributes of a Pod's containers (userID, groupID, Linux capabilities, etc.). In your Operator implementation, you may decide to create Custom Resource Pods using
certain settings for the securityContext. Surface these settings through 'platform-as-code/securitycontext' annotation
on the CRD. The value of this annotation should be the name of a ConfigMap that contains the security context attributes defined as the data values. Make sure to include this ConfigMap in your Operator's Helm chart.
By surfacing the security context information in this way, it will be possible for the DevOps engineers and
Application developers to find out the security context with which Custom Resource Pods are going to run.


## 9) Evaluate Service Account needs for Custom Resource Pods

Your Custom Resource's Pods may need to run with specific Service account. If that is the case, one
of the decisions you will need to make is whether that Service account should be provided by 
users of your Custom Resource. If so, you will need to provide an attribute in Custom Resource Spec
definition to define this Service account. Alternatively, if the custom controller is hard coding
the Service account in the Pod Spec that it creates, you will need to surface this information
for your users (check guideline #20 for details).


# Custom Controllers

## 10) Make Custom Controllers namespace aware

Your Operator should support creating resources within different namespaces rather than just in the default namespace. This will allow your Operator to support multitenancy through namespaces.


## 11) Set OwnerReferences for underlying resources owned by your Custom Resource

An Operator will typically create one or more native Kubernetes resources as part of instantiating a Custom Resource instance. Set the OwnerReference attribute of such underlying resources to the Custom Resource instance that is
being created. OwnerReferences form the key attribute for correct garbage collection of custom resources. OwnerReferences also help with finding runtime composition tree of your custom resource instances.


## 12) Decide Custom Resource Metrics Collection strategy

Plan for metrics collection of custom resources managed by your Operator. This information is useful for understanding effect of various actions on your custom resources over time and improving traceability. 
For example, [this MySQL Operator](https://github.com/oracle/mysql-operator/) 
collects metrics such as how many clusters were created. One option to collect such metrics 
is to build the metrics collection inside your custom controller as done by the MySQL Operator.
Another option is to leverage Kubernetes Audit Logs for this purpose. 
Then you can use external tooling like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) 
to build the required metrics. Once metrics are collected, you should consider exposing the collected metrics 
in Prometheus format.


# Operator packaging


## 13) Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * All Custom Resource Definitions for Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * ConfigMaps corresponding to Platform-as-Code annotations that you have added on your Custom Resource Definition (CRD).


## 14) Register CRDs as YAML Spec in Helm chart rather than in Operator code

Installing CRD requires Cluster-scope permission. If the CRD registration is done as YAML manifest, then it is possible to separate CRD registration from the Operator Pod deployment. CRD registration
can be done by DevOps engineers while Operator Pod deployment can be done by a non-admin user. 
It is then possible to deploy the Operator in different namespaces with different customizations.
On the other hand, if CRD registration is done as part of your Operator code then the deployment of the Operator Pod will need Cluster-scope permissions.
Another reason to register CRD as YAML is because kube-openapi validation can be defined as part of it.


## 15) Add crd-install Helm hook annotation on your CRD YAML

Helm defines crd-install hook that directs Helm to install CRDs first before installing rest of your
Helm chart that might refer to the Custom Resources defined by the CRDs. 
This is important as otherwise the Custom Resources defined in your chart won't be able to be
installed in your cluster.

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: moodles.moodlecontroller.kubeplus
    annotations:
      helm.sh/hook: crd-install
```

## 16) Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML

Your Custom Resource Spec definitions will contain different properties and they may have some
domain-specific validation requirements. Kubernetes 1.13 onwards you will be able to use 
OpenAPI v3 schema to define validation requirements for your Custom Resource Spec. For instance,
below is an example of adding validation rules for our sample Postgres CRD. The rules define that
the Postgres Custom Resource Spec properties of 'databases' and 'users' should be of type Array
and that every element of this array should be of type String. Once such validation rules are defined,
Kubernetes will reject any Custom Resource instance creation that does not satisfy these requirements
in their Spec.

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


# Operator and underlying resource configuration

## 17) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables

An Operator generally needs to take configuration parameter as inputs 
for the underlying resource that it is managing through its custom resource such as a database.
We have seen three different approaches being used towards this in the community: using ConfigMaps, using Annotations, or using Spec definition itself. Any of these approaches should be fine based on your Operator design. 
It is also possible that you may end up using multiple approaches, such as a ConfigMap with its name specified in the 
Custom Resource Spec definition.

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).


## 18) Use Helm chart or ConfigMap for Operator configurables

Typically Operators will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have created Helm Chart for your Operator then use values YAML file to specify
such parameters. If not, use ConfigMap for this purpose. If you choose to use ConfigMap then make sure the
name of the ConfigMap that your Operator expects is well-documented so that DevOps Engineers can create
this ConfigMap.


# Usability

## 19) Consider to use kubectl as the primary interaction mechanism

Custom resources introduced by your Operator will naturally work with kubectl.
However, there might be operations that you want to support for which the declarative nature of custom resources
is not appropriate. An example of such an action is historical record of how Postgres Custom Resource has evolved over time
that might be supported by the Postgres Operator. Such an action does not fit naturally into the declarative
format of custom resource definition. For such actions, we encourage you to consider using Kubernetes
extension mechanisms of Aggregated API servers and Custom Sub-resources. These mechanisms 
will allow you to continue using kubectl as the primary interaction point for your Operator.
Refer to [this blog post](https://medium.com/@cloudark/comparing-kubernetes-api-extension-mechanisms-of-custom-resource-definition-and-aggregated-api-64f4ca6d0966) to learn more about them. Before considering to introduce new CLI for your Operator, validate if you can use these mechanisms instead. 


## 20) Add Platform-as-Code annotations on your CRD YAML

[Platform-as-Code annotations](https://github.com/cloud-ark/kubeplus#platform-as-code-annotations) provides an approach to package and surface information about Custom Resources.
Following annotations are available - 'usage', 'composition', 'pdb-minavailable', 
'platform-as-code/pdb-maxunavailable', 'platform-as-code/cpu-requests', 'platform-as-code/cpu-limits', 'platform-as-code/mem-requests', 'platform-as-code/mem-limits', 'platform-as-code/securitycontext', 'platform-as-code/serviceaccount'.
The 'usage' annotation should be used to define how-to use guide of the Custom Resource.
The 'composition' annotation should be used to specify the underlying Kubernetes resources that will be created as part of managing a Custom Resource instance. 
The values of all the annotations, except the 'composition' annotation, are names of ConfigMaps with appropriate data.
An example of this can be seen for our sample Moodle Custom Resource Definition below:

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

This information is useful for application developers when figuring out how to use your Operator and its Custom Resources. Externalizing information like that available in the 'composition' annotation makes it possible to build tools like [kubediscovery](https://github.com/cloud-ark/kubediscovery) that show Object composition tree for custom resource instances built leveraging this information.

Following annotations - 'platform-as-code/cpu-requests', 'platform-as-code/cpu-limits', 'platform-as-code/mem-requests', 
'platform-as-code/mem-limits' will help the DevOps engineers to understand the resource requirements of the Custom Resource Pods. This in turn will enable them to define the per namespace resource quotas for different Operators. 

Following annotations - 'platform-as-code/pdb-maxunavailable' or 'platform-as-code/pdb-minavailable' 
will help Application developers understand the availability guarantee for Custom Resources provided by your Operator.


# Documentation guidelines

## 21) Document how your Operator uses namespaces

For Operator developers it is critical to consider how their Operator works with namespaces. Typically, an Operator can be installed in one of the following configurations:

  * Operator runs in the default namespace and Custom Resource instances are created in the default namespace.

  * Operator runs in the default namespace but Custom Resource instances can be created in non-default namespaces.

  * Operator runs in a non-default namespace and Custom Resource instances can be created in that namespace.

Given these options, it will help consumers of your Operator if there is a clear documentation of how namespaces are used by your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## 22) Document Service Account needs of your Operator

Your Operator may be need to use a specific service account with specific permissions. Clearly document the service account needs of your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## 23) Document naming convention and labels to be used with your Custom Resources

You may have special requirements for naming your custom resource instances or some of their
Spec properties. Similarly you may have requirements related to the labels that need to be added on them. Document this information in the ConfigMap corresponding to the 'usage' platform-as-code annotation on the CRD.








