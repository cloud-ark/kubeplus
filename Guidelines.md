# Kubernetes Operator Maturity Model

Kubernetes Operators enable running third-party softwares directly on Kubernetes.
Various Operators are being built today for variety of softwares such as 
MySQL, Postgres, Cassandra, Airflow, Kafka, Prometheus, Moodle, Wordpress, etc.
Technically, a Kubernetes Operator consists of one or more Kubernetes Custom Resources and their associated Custom Controllers. A Custom Resource provides declarative model to specify the desired state of a resource
that the associated Custom Controller manages. While an individual Operator is typically focused on 
managing some domain-specific workflows in Kubernetes-native manner, increasingly there are setups where more
than one Operators are installed on a cluster.

Below we define a Operator maturity model which is intended to capture the 
wide range of Operator setups that are seen in enterprises today. 
We also present guidelines that an Operator developer should follow
in order to make their Operator compliant towards each maturity level.
 
### 1) Application developer usability

This maturity level defines the requirements related to the usability of Custom Resources of an Operator. These include things such as - how to use Custom Resources to define different workflow actions, how to find out runtime information about Custom Resources, ability of application developers to discover the capabilities offered by the Operator including any implementation-level assumptions made by the Operator developer, etc.


### 2) Multi-Operator interoperability guarantees

This maturity level identifies the requirements related to using your Operator alongside other Operators in a cluster. This includes things such as enabling application developers to specify resource requests and limits for your Custom Resources, defining Custom Resource affinity policies with other Custom Resources, defining Custom Resource node co-location policies, etc.

### 3) Multi-tenant guarantees

This maturity level identifies requirements related to using an Operator
in creating multi-tenant stacks consisting of different Custom Resources. 

### 4) Kubernetes Distribution and Cloud provider independence

This maturity level identifies the requirements related to Operator packaging and installation. The goal here should be to create Operator packaging and installation artifacts such that they are independent of Kubernetes distribution or Cloud provider. This will enable the Operator to be installed on any Kubernetes cluster - on-prem or public clouds.

The maturity model is intended to be used to calibrate an Operator's readiness
to work in increasingly complex scenarios. You can use it as a guiding framework when thinking about your
next Operator design. 


## Guidelines


### 1) Application developer usability

[Design Custom Resource as a declarative API and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#design-operator-with-declarative-apis-and-avoid-inputs-as-imperative-actions)

[Make Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#make-custom-resource-type-definitions-compliant-with-kube-openapi)

[Use ConfigMap or Custom Resource Annotation or Custom Resource Spec definition for underlying resource configuration](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#use-configmap-or-custom-resource-annotation-or-custom-resource-spec-definition-for-underlying-resource-configuration)

[Define PodDisruptionBudget for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-poddisruptionbudget-for-custom-resources)

[Add Platform-as-Code annotations on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#add-platform-as-code-annotations-on-your-crd-yaml)

[Consider to use kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#consider-to-use-kubectl-as-the-primary-interaction-mechanism)

[Document Service Account needs of your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-service-account-needs-of-your-operator)

[Document naming convention and labels to be used with your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-naming-convention-and-labels-to-be-used-with-your-custom-resources)


### 2) Multi-Operator interoperability guarantees

[Define Resource limits and Resource requests for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-resource-limit-and-resource-requests-for-custom-resources)

[Define Custom Resource Node Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-node-affinity-rules)

[Define Custom Resource Pod Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-pod-affinity-rules)

[Define Custom Resource Anti-Affinity rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-anti-affinity-rules)

[Define Custom Resource Taint Toleration rules](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-taint-toleration-rules)

[Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[Decide Custom Resource Metrics Collection strategy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#decide-custom-resource-metrics-collection-strategy)


### 3) Multi-tenant guarantees

[Define SecurityContext for Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-securitycontext-for-custom-resources)

[Evaluate Service Account needs for Custom Resource Pods](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#evaluate-service-account-needs-for-custom-resource-pods)

[Make Custom Controllers Namespace aware](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#make-custom-controllers-namespace-aware)

[Document how your Operator uses namespaces](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#document-how-your-operator-uses-namespaces)


### 4) Kubernetes Distribution and Cloud provider independence

[Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#package-operator-as-helm-chart)

[Register CRDs as YAML Spec in Helm chart rather than in Operator code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#register-crds-as-yaml-spec-in-helm-chart-rather-than-in-operator-code)

[Add crd-install Helm hook annotation on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#add-crd-install-helm-hook-annotation-on-your-crd-yaml)

[Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition-yaml)

[Use Helm chart or ConfigMap for Operator configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#use-helm-chart-or-configmap-for-operator-configurables)





# Detail Guidelines

## Design Operator with declarative API/s and avoid inputs as imperative actions

A declarative API is one in which you specify the desired state of your custom resource using the Custom Resource Spec definition. Prefer declarative specification over any imperative actions in Custom Resource Spec Type definition. Custom controller code should be written such that it reconciles the current state of the underlying software with the desired state by performing diff of the current state with the desired state. For example, when writing a Postgres Operator, the custom controller should be written to perform diff of the existing value of ‘users’ with the desired value of ‘users’ based on the received desired state and perform the required actions (such as adding new users, deleting current users, etc.). An example where underlying imperative actions are exposed in the Spec is this 
[MySQL Backup Custom Resource Spec](https://github.com/oracle/mysql-operator/blob/master/examples/backup/backup.yaml#L7). Here the fact that MySQL Backup is done using mysqldump tool is exposed in the Spec.
In our view such internal details should not be exposed in the Spec as it prevents Custom Resource Type definition 
to evolve independently without affecting its users.

## Make Custom Resource Type definitions compliant with Kube OpenAPI

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

## Define Resource limit and Resource requests for Custom Resources

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
Custom Resource YAML (CRD YAML) ([check guideline for Platform-as-Code annotations for details]
(https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#add-platform-as-code-annotations-on-your-crd-yaml)).


## Define PodDisruptionBudget for Custom Resources

Kubernetes provides mechanism of [Pod Disruption Budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) (PDB) that can be used to define the disruption tolerance for Pods. Specifically, two
fields are provided - 'minAvailable' and 'maxUnavailable'. minAvailable is the minimum number of Pods that 
should be always running in a cluster. maxUnavailable is complementary and defines the maximum number of Pods
that can be unavailable in a cluster. These two fields provide a way to control the availability of Pods in a cluster.
They ensure that minimum number of Pods will always be present. When implementing the controller for your Custom Resource, carefully consider such availability requirements for your Custom Resource instance's Pods. If you do decide to implement PDB for your Custom Resource Pods, you will have to decide whether you want to expose this control to your Custom Resource 
users. If yes, then ensure that the Custom Resource Spec definition has a field to specify a disruption budget.
If on the other hand you decide to hard code this choice in your custom controller implementation then
surface it to your Operator users through Custom Resource Definition YAML (CRD YAML)
([check guideline for Platform-as-Code annotations for details]
(https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#add-platform-as-code-annotations-on-your-crd-yaml)).


## Define Custom Resource Node Affinity rules

Kubernetes provides mechanism of [Pod Node Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)
This mechanism enables defining the specific nodes on which a Pod should run.
The way this is achieved is by providing a set of labels on the Pod Spec that are matched by the scheduler 
with the labels on the nodes when making the scheduling decision. If your Custom Resource Pods need to be
subject to this scheduling constraint then you will need to define the Custom Resource Spec to allow input of such labels.


## Define Custom Resource Pod Affinity rules

Kubernetes provides mechanism of [Pod Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which enables defining Pod scheduling rules based on labels of other Pods that are running on a node. 
Consider if your Custom Resource Pods need to be provided with such affinity rules corresponding
to other Custom Resources from same or other Operator. If so, provide an attribute in your Custom Resource
Spec definition where such rules can be specified. 


## Define Custom Resource Anti-Affinity rules

Kubernetes also provides mechanism of [Pod Anti-Affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) which are opposite of Pod affinity rules. 
Consider if your Custom Resource Pods need to be provided with such anti-affinity rules corresponding to 
other Custom Resoures from other Operators. If so, provide an attribute in your Custom Resource
Spec definition where such rules can be specified. 


## Define Custom Resource Taint Toleration rules

Kubernetes provides mechanism of [taints and tolerations](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) to restrict scheduling of Pods on certain nodes. If you want your Custom Resource pods
to be able to tolerate the taints on a node, then define an attribute in your Custom Resource Spec definition
where such tolerations can be specified. 


## Define SecurityContext for Custom Resources

Kubernetes provides mechanism of [SecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) that can be used to define the security attributes of a Pod's containers (userID, groupID, Linux capabilities, etc.). In your Operator implementation, you may decide to create Custom Resource Pods using
certain settings for the securityContext. Surface these settings through 'platform-as-code/securitycontext' annotation
on the CRD. The value of this annotation should be the name of a ConfigMap that contains the security context attributes defined as the data values. Make sure to include this ConfigMap in your Operator's Helm chart.
By surfacing the security context information in this way, it will be possible for the DevOps engineers and
Application developers to find out the security context with which Custom Resource Pods are going to run.


## Evaluate Service Account needs for Custom Resource Pods

Your Custom Resource's Pods may need to run with specific Service account. If that is the case, one
of the decisions you will need to make is whether that Service account should be provided by 
users of your Custom Resource. If so, you will need to provide an attribute in Custom Resource Spec
definition to define this Service account. Alternatively, if the custom controller is hard coding
the Service account in the Pod Spec that it creates, you will need to surface this information
for your users (check guideline #20 for details).

## Make Custom Controllers namespace aware

Your Operator should support creating resources within different namespaces rather than just in the default namespace. This will allow your Operator to support multi-tenancy through namespaces.


## Set OwnerReferences for underlying resources owned by your Custom Resource

An Operator will typically create one or more native Kubernetes resources as part of instantiating a Custom Resource instance. Set the OwnerReference attribute of such underlying resources to the Custom Resource instance that is
being created. OwnerReferences form the key attribute for correct garbage collection of custom resources. OwnerReferences also help with finding runtime composition tree of your custom resource instances.


## Decide Custom Resource Metrics Collection strategy

Plan for metrics collection of custom resources managed by your Operator. This information is useful for understanding effect of various actions on your custom resources over time and improving traceability. 
For example, [this MySQL Operator](https://github.com/oracle/mysql-operator/) 
collects metrics such as how many clusters were created. One option to collect such metrics 
is to build the metrics collection inside your custom controller as done by the MySQL Operator.
Another option is to leverage Kubernetes Audit Logs for this purpose. 
Then you can use external tooling like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) 
to build the required metrics. Once metrics are collected, you should consider exposing the collected metrics 
in Prometheus format.


## Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * All Custom Resource Definitions for Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * ConfigMaps corresponding to Platform-as-Code annotations that you have added on your Custom Resource Definition (CRD).


## Register CRDs as YAML Spec in Helm chart rather than in Operator code

Installing CRD requires Cluster-scope permission. If the CRD registration is done as YAML manifest, then it is possible to separate CRD registration from the Operator Pod deployment. CRD registration
can be done by DevOps engineers while Operator Pod deployment can be done by a non-admin user. 
It is then possible to deploy the Operator in different namespaces with different customizations.
On the other hand, if CRD registration is done as part of your Operator code then the deployment of the Operator Pod will need Cluster-scope permissions.
Another reason to register CRD as YAML is because kube-openapi validation can be defined as part of it.


## Add crd-install Helm hook annotation on your CRD YAML

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

## Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML

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


## Use ConfigMap or Custom Resource Annotation or Custom Resource Spec definition for underlying resource configuration

An Operator generally needs to take configuration parameter as inputs 
for the underlying resource that it is managing through its custom resource such as a database.
We have seen three different approaches being used towards this in the community: using ConfigMaps, using Annotations, or using Spec definition itself. Any of these approaches should be fine based on your Operator design. 
It is also possible that you may end up using multiple approaches, such as a ConfigMap with its name specified in the 
Custom Resource Spec definition.

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).


## Use Helm chart or ConfigMap for Operator configurables

Typically Operators will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have created Helm Chart for your Operator then use values YAML file to specify
such parameters. If not, use ConfigMap for this purpose. If you choose to use ConfigMap then make sure the
name of the ConfigMap that your Operator expects is well-documented so that DevOps Engineers can create
this ConfigMap.


## Consider to use kubectl as the primary interaction mechanism

Custom resources introduced by your Operator will naturally work with kubectl.
However, there might be operations that you want to support for which the declarative nature of custom resources
is not appropriate. An example of such an action is historical record of how Postgres Custom Resource has evolved over time
that might be supported by the Postgres Operator. Such an action does not fit naturally into the declarative
format of custom resource definition. For such actions, we encourage you to consider using Kubernetes
extension mechanisms of Aggregated API servers and Custom Sub-resources. These mechanisms 
will allow you to continue using kubectl as the primary interaction point for your Operator.
Refer to [this blog post](https://medium.com/@cloudark/comparing-kubernetes-api-extension-mechanisms-of-custom-resource-definition-and-aggregated-api-64f4ca6d0966) to learn more about them. Before considering to introduce new CLI for your Operator, validate if you can use these mechanisms instead. 


## Add Platform-as-Code annotations on your CRD YAML

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


## Document how your Operator uses namespaces

For Operator developers it is critical to consider how their Operator works with namespaces. Typically, an Operator can be installed in one of the following configurations:

  * Operator runs in the default namespace and Custom Resource instances are created in the default namespace.

  * Operator runs in the default namespace but Custom Resource instances can be created in non-default namespaces.

  * Operator runs in a non-default namespace and Custom Resource instances can be created in that namespace.

Given these options, it will help consumers of your Operator if there is a clear documentation of how namespaces are used by your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## Document Service Account needs of your Operator

Your Operator may be need to use a specific service account with specific permissions. Clearly document the service account needs of your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## Document naming convention and labels to be used with your Custom Resources

You may have special requirements for naming your custom resource instances or some of their
Spec properties. Similarly you may have requirements related to the labels that need to be added on them. Document this information in the ConfigMap corresponding to the 'usage' platform-as-code annotation on the CRD.








