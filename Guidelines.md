# Kubernetes Operator Guidelines to enable Multi-Operator Platform Stacks (Platform-as-Code)

Kubernetes Operators extend Kubernetes API to run third-party softwares on Kubernetes.
Various Operators are being built today for variety of softwares such as 
MySQL, Postgres, Airflow, Redis, MongoDB, Kafka, Prometheus, Logstash, Moodle, Wordpress, Odoo, etc.
We are seeing a new trend where Custom Resources introduced by different Operators are used to create custom application [platforms as Code](https://cloudark.io/platform-as-code). Such platforms are Kubernetes-native, they don't have any platform vendor lock-in, and they can be created/re-created on any Kubernetes cluster.

Towards building such platforms various Operators and their Custom Resources need to provide a consistent usage experience. We present below guidelines that ensure such an experience to end users (Cluster Administrators and Application Developers). Operators that are developed following these guidelines provide ease of installation, management, and discovery to end users who are creating application platforms leveraging them.

Check out [this post](https://medium.com/@cloudark/analysis-of-open-source-kubernetes-operators-f6be898f2340) about our analysis of more than 100 open source Operators for their conformance to some of these guidelines.

## Design guidelines

[1) Design Operator with declarative API/s and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#1-design-operator-with-declarative-apis-and-avoid-inputs-as-imperative-actions)

[2) Consider to use kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#2-consider-to-use-kubectl-as-the-primary-interaction-mechanism)

[3) Decide Custom Resource Metrics Collection strategy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#3-decide-custom-resource-metrics-collection-strategy)

[4) Register CRDs as YAML Spec rather than in Operator code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#4-register-crds-as-yaml-spec-rather-than-in-operator-code)


## Implementation guidelines

[5) Make Operator namespace aware](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#5-make-operator-namespace-aware)

[6) Make Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#6-make-custom-resource-type-definitions-compliant-with-kube-openapi)

[7) Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#7-define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition-yaml)

[8) Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#8-set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[9) Use Helm chart or ConfigMap for Operator configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#9-use-helm-chart-or-configmap-for-operator-configurables)

[10) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#10-use-configmap-or-annotation-or-spec-definition-for-custom-resource-configurables)


## Packaging guidelines

[11) Generate Kube OpenAPI Spec for your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#11-generate-kube-openapi-spec-for-your-custom-resources)

[12) Add Platform-as-Code annotations on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#12-add-platform-as-code-annotations-on-your-crd-yaml)

[13) Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#13-package-operator-as-helm-chart)


[14) Add crd-install Helm hook annotation on your CRD YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#14-add-crd-install-helm-hook-annotation-on-your-crd-yaml)



## Documentation guidelines

[15) Document how your Operator uses namespaces](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#15-document-how-your-operator-uses-namespaces)

[16) Document Service Account needs of your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#16-document-service-account-needs-of-your-operator)

[17) Document naming convention and labels to be used with your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#17-document-naming-convention-and-labels-to-be-used-with-your-custom-resources)


# Design guidelines

## 1) Design Operator with declarative API/s and avoid inputs as imperative actions

A declarative API allows you to declare or specify the desired state of your custom resource. Prefer declarative state over any imperative actions in Custom Resource Spec Type definition. Custom controller code should be written such that it reconciles the current state with the desired state by performing diff of the current state with the desired state. This enables end users to use your custom resources just like any other Kubernetes resources with declarative state based inputs. For example, when writing a Postgres Operator, the custom controller should be written to perform diff of the existing value of ‘users’ with the desired value of ‘users’ based on the received desired state and perform the required actions (such as adding new users, deleting current users, etc.).

Note that the diff-based implementation approach for custom controllers is essentially an extension of
the level-triggered approach recommended in the [general guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md) for developing Kubernetes controllers.

An example where underlying imperative actions are exposed in the Spec is this 
[MySQL Backup Custom Resource Spec](https://github.com/oracle/mysql-operator/blob/master/examples/backup/backup.yaml#L7).
Here the fact that MySQL Backup is done using mysqldump tool is exposed in the Spec.
In our view such internal details should not be exposed in the Spec as it prevents Custom Resource Type definition 
to evolve independently without affecting its users.


## 2) Consider to use kubectl as the primary interaction mechanism

Custom resources introduced by your Operator will naturally work with kubectl.
However, there might be operations that you want to support for which the declarative nature of custom resources
is not appropriate. An example of such an action is historical record of how Postgres Custom Resource has evolved over time
that might be supported by the Postgres Operator. Such an action does not fit naturally into the declarative
format of custom resource definition. For such actions, we encourage you to consider using Kubernetes
extension mechanisms of Aggregated API servers and Custom Sub-resources. These mechanisms 
will allow you to continue using kubectl as the primary interaction point for your Operator.
Refer to [this blog post](https://medium.com/@cloudark/comparing-kubernetes-api-extension-mechanisms-of-custom-resource-definition-and-aggregated-api-64f4ca6d0966) to learn more about them.
Before considering to introduce new CLI for your Operator, validate if you can use these mechanisms instead. 


## 3) Decide Custom Resource Metrics Collection strategy

Plan for metrics collection of custom resources managed by your Operator. This information is useful for understanding effect of various actions on your custom resources over time and improving traceability. 
For example, [this MySQL Operator](https://github.com/oracle/mysql-operator/) 
collects metrics such as how many clusters were created. One option to collect such metrics 
is to build the metrics collection inside your custom controller as done by the MySQL Operator.
Another option is to leverage Kubernetes Audit Logs for this purpose. 
Then you can use external tooling like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) 
to build the required metrics. Once metrics are collected, you should consider exposing the collected metrics 
in Prometheus format.


## 4) Register CRDs as YAML Spec rather than in Operator code

Registering CRDs as YAML spec rather than in your Operator code has following advantage.
Installing CRD requires Cluster-scope permission. If the CRD registration is done as YAML manifest, then it is possible to separate CRD registration from the Operator Pod deployment. CRD registration
can be done by Cluster administrator while Operator Pod deployment can be done by a non-admin user. 
It is then possible to deploy the Operator in different namespaces with different customizations.
On the other hand, if CRD registration is done as part of your Operator code then the deployment of the Operator Pod will need Cluster-scope permissions.
Another reason to register CRD as YAML is because kube-openapi validation can be defined as part of it.


# Implementation guidelines


## 5) Make Operator namespace aware

Your Operator should support creating resources within different namespaces rather than just in the default namespace. This will allow your Operator to support use-cases where multitenancy through namespaces is sufficient (these are called 'soft' multitenancy use-cases).


## 6) Make Custom Resource Type definitions compliant with Kube OpenAPI

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

## 7) Define Custom Resource Spec Validation rules as part of Custom Resource Definition YAML

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


## 8) Set OwnerReferences for underlying resources owned by your Custom Resource

An Operator will typically create one or more native Kubernetes resources, such as Deployment, Service, Secret, etc., as part of instantiating a Custom Resource instance. Here this custom resource is the owner of its underlying native resources that were created by the Operator. OwnerReferences are key for correct garbage collection of custom resources. OwnerReferences also help with finding runtime composition tree of your custom resource instances.


## 9) Use Helm chart or ConfigMap for Operator configurables

Typically Operators will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have created Helm Chart for your Operator then use values YAML file to specify
such parameters. If not, use ConfigMap for this purpose. If you choose to use ConfigMap then make sure the
name of the ConfigMap that your Operator expects is well-documented so that DevOps Engineers can create
this ConfigMap. This guideline ensures that Kubernetes Administrators can interact and use the Operator using Kubernetes native's interfaces.


## 10) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables

An Operator generally needs to take configuration parameter as inputs 
for the underlying resource that it is managing through its custom resource such as a database.
We have seen three different approaches being used towards this in the community: using ConfigMaps, using Annotations, or using Spec definition itself. Any of these approaches should be fine based on your Operator design. 
You may end up using multiple approaches such as a ConfigMap with its name specified in the 
Custom Resource Spec definition.

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).


# Packaging guidelines

## 11) Generate Kube OpenAPI Spec for your Custom Resources

We have developed a [tool](https://github.com/cloud-ark/kubeplus/tree/master/openapi-spec-generator) that you can use for generating Kube OpenAPI Spec for your custom resources. 
It wraps code available in [kube-openapi repository](https://github.com/kubernetes/kube-openapi) 
in an easy to use script. You can use this tool to generate OpenAPI Spec for your custom resources.
The generated Kube OpenAPI Spec documentation for sample Postgres custom resource is 
[here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/openapispec.json).
The OpenAPI Spec for your Operator provides a single place where documentation is available for the entire Type
definition hierarchy for the custom resources defined by your Operator.


## 12) Add Platform-as-Code annotations on your CRD YAML

[Platform-as-Code annotations](https://github.com/cloud-ark/kubeplus#platform-as-code-annotations) are a standard way to package information about Custom Resources.
The 'usage' annotation should be used to define how-to use guide of a Custom Resource.
The 'constants' annotation should be used to define Operator's implementation choices and assumption.
The 'composition' annotation should be used to specify the underlying Kubernetes resources that will be created as part of managing a Custom Resource instance. The values of the first three annotations are names of ConfigMaps with appropriate data.
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


## 13) Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * All Custom Resource Definitions for Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * ConfigMaps corresponding to Platform-as-Code annotations that you have added on your Custom Resource Definition (CRD).


## 14) Add crd-install Helm hook annotation on your CRD YAML

Helm defines crd-install hook that allows Helm to install CRDs first before installing rest of your
Helm chart that might refer to the Custom Resources defined by the CRDs. 
This is important as otherwise the Custom Resources defined in your chart won't be recognized in your cluster.



# Documentation guidelines

## 15) Document how your Operator uses namespaces

For Operator developers it is critical to consider how their Operator works with namespaces. Typically, an Operator can be installed in one of the following configurations:

  * Operator runs in the default namespace and Custom Resource instances are created in the default namespace.

  * Operator runs in the default namespace but Custom Resource instances can be created in non-default namespaces.

  * Operator runs in a non-default namespace and Custom Resource instances can be created in that namespace.

Given these options, it will help consumers of your Operator if there is a clear documentation of how namespaces are used by your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## 16) Document Service Account needs of your Operator

Your Operator may be need to use some specific service account. Moreover, the service account may need to be granted specific permissions. Clearly document the service account needs of your Operator. Include this information in the ConfigMap that you will add for the 'usage' platform-as-code annotation on the CRD.


## 17) Document naming convention and labels to be used with your Custom Resources

You may have special requirements for naming your custom resource instances or some of their
Spec properties. Similarly you may have requirements related to the labels that need to be added on them. Document this information with in the ConfigMap corresponding to the 'usage' platform-as-code annotation on the CRD.








