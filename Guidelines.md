# Kubernetes Operator Development Guidelines for Discoverability and Interoperability

Kubernetes Operators extend Kubernetes API to manage third-party software as native Kubernetes objects.
Various Operators are being built today for variety of softwares such as 
MySQL, Postgres, Airflow, Redis, MongoDB, Kafka, Prometheus, Logstash, Moodle, Wordpress, Odoo, etc.
to run on Kubernetes. Consequently, we are seeing a new trend where fit-for-purpose application platforms 
are being created by composing multiple such Operators together on a Kubernetes cluster.
The Custom Resources introduced by different Operators can be used to create custom application Platforms as Code.

While working in this space we have observed following challenges when using multiple Operators together.

  * CLI overload: Some of the Operators introduce new CLIs. 
Usability becomes an issue when end users have to learn multiple CLIs to use more than one Operators in a cluster.

  * API diversity: When using multiple Operators, it becomes a challenge for application developers 
to discover the capabilities of the various Custom Resources available in the Cluster —
what all Custom Resources are available? what are their attributes? how to use them?
Application developers can go to each Operator’s documentation to find this information, 
but this is not a user-friendly approach.

  * Interoperability: Another challenge is how different Custom Resources work together. 
For instance, a MySQL Custom Resource and a Backup Custom Resource both may work with Volumes. 
How to ensure that both these Custom Resources are using the same Volume in their operations?

There is much diversity in how Operators are implemented today. Below presented guidelines are our attempt
to formalize a basic set of rules that Operator developers can follow while developing their Operators.
We have developed these guidelines after studying existing community Operators and building Operators
and related tools ourselves. These guidelines are developed with the goal to ensure that Operators 
will provide consistent usage experience to end users 
(cluster administrators and application developers). Specifically, cluster admins will be able to 
easily compose multiple Operators together to form a platform stack;
and application developers will be able to discover and consume Custom Resources from different
Operators effortlessly.

The guidelines are divided into four sections - design guidelines, implementation guidelines, packaging guidelines
and documentation guidelines.


## Design guidelines

[1) Design your Operator with declarative API/s and avoid inputs as imperative actions](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#1-design-your-operator-with-declarative-apis-and-avoid-inputs-as-imperative-actions)

[2) Consider to use kubectl as the primary interaction mechanism](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#2-consider-to-use-kubectl-as-the-primary-interaction-mechanism)

[3) Decide your Custom Resource Metrics Collection strategy](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#3-decide-your-custom-resource-metrics-collection-strategy)

[4) Register CRDs as part of Operator Helm chart rather than in code](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#4-register-crds-as-part-of-operator-helm-chart-rather-than-in-code)

[5) Make Operator ETCD dependency configurable](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#5-make-operator-etcd-dependency-configurable)

[6) Make Operator namespace aware](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#6-make-operator-namespace-aware)


## Implementation guidelines

[7) Set OwnerReferences for underlying resources owned by your Custom Resource](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#7-set-ownerreferences-for-underlying-resources-owned-by-your-custom-resource)

[8) Use Helm chart or ConfigMap for Operator configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#8-use-helm-chart-or-configmap-for-operator-configurables)

[9) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#9-use-configmap-or-annotation-or-spec-definition-for-custom-resource-configurables)

[10) Declare underlying resources created by Custom Resource as Annotation on CRD registration YAML](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#10-declare-underlying-resources-created-by-custom-resource-as-annotation-on-crd-registration-yaml)

[11) Make your Custom Resource Type definitions compliant with Kube OpenAPI](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#11-make-your-custom-resource-type-definitions-compliant-with-kube-openapi)

[12) Define Custom Resource Spec Validation rules as part of Custom Resource Definition](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#12-define-custom-resource-spec-validation-rules-as-part-of-custom-resource-definition)


## Packaging guidelines

[13) Generate Kube OpenAPI Spec for your Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#13-generate-kube-openapi-spec-for-your-custom-resources)

[14) Package Operator as Helm Chart](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#14-package-operator-as-helm-chart)

## Documentation guidelines

[15) Document how your Operator uses namespaces](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#15-document-how-your-operator-uses-namespaces)

[16) Document Service Account usage of your Operator](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#16-document-service-account-usage-of-your-operator)

[17) Document naming convention and labels needed to be used with Custom Resources](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md#17-document-naming-convention-and-labels-needed-to-be-used-with-custom-resources)


# Design guidelines

## 1) Design your Operator with declarative API/s and avoid inputs as imperative actions

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
So before considering to introduce new CLI for your Operator, validate if you can use these mechanisms instead. 


## 3) Decide your Custom Resource Metrics Collection strategy

Plan for metrics collection of custom resources managed by your Operator. This information is useful for understanding effect of various actions on your custom resources over time and improving traceability. 
For example, [this MySQL Operator](https://github.com/oracle/mysql-operator/) 
collects metrics such as how many clusters were created. One option to collect such metrics 
is to build the metrics collection inside your custom controller as done by the MySQL Operator.
Another option is to leverage Kubernetes Audit Logs for this purpose. 
Then you can use external tooling like [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) 
to build the required metrics. Separately, you can consider exposing the collected metrics 
in Prometheus format as well.


## 4) Register CRDs as part of Operator Helm chart rather than in code

Registering CRDs as part of Operator Helm Chart rather than in your Operator code has following advantages:

  * All installation artifacts and dependencies are in one place — the Operator’s Helm Chart.

  * It is easy to modify and/or evolve the CRD by just updating the Chart.


## 5) Make Operator ETCD dependency configurable

If your Operator needs ETCD for its storage then its best to make this dependency configurable through your 
Operator Helm Chart. This will allow Platform engineers installing your Operator along with other Operators to
decide the best way to use ETCD. If there are multiple Operators needing ETCD then by making this configurable
you will allow Platform Engineer to decide whether they want to provide separate ETCD instance to each Operator
or use a shared ETCD instance between all of them. It is possible Platform Engineer may use ETCD Operator
to provision ETCD instances. Your Operator code should not depend on how ETCD instance is made available to it.


## 6) Make Operator namespace aware

Your Operator should support creating resources within different namespaces rather than just in the default namespace. This will allow your Operator to support multi-tenant usecases.


# Implementation guidelines

## 7) Set OwnerReferences for underlying resources owned by your Custom Resource

A custom resource instance will typically create one or more other Kubernetes resource instances, such as Deployment, Service, Secret etc., as part of its instantiation. Here this custom resource is the owner of its underlying resources that it manages. Custom controller should be written to set OwnerReference on such managed Kubernetes resources. They are key for correct garbage collection of custom resources. OwnerReferences also help with finding composition tree of your custom resource instances. 

Here are some examples of Operators that use OwnerReferences: [Etcd Operator](https://github.com/coreos/etcd-operator/blob/master/pkg/cluster/cluster.go#L351),
[Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/controller.go#L508), and 
[MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/pkg/resources/services/service.go#L34).


## 8) Use Helm chart or ConfigMap for Operator configurables

Typically Operators will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have created Helm Chart for your Operator then use values YAML file to specify
such parameters. If not, use ConfigMap for this purpose. This guideline ensures that Kubernetes Administrators
can interact and use the Operator using Kubernetes native's interfaces.


## 9) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables

An Operator generally needs to take configuration parameter as inputs 
for the underlying resource that it is managing through its custom resource such as a database.
We have seen three different approaches being used towards this in the community
- using ConfigMaps, using Annotations, or using Spec definition itself. 
Any of these approaches should be fine based on your Operator design. 

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).

Similar to guideline #6, this guideline ensures that application developers can interact and use Custom Resources using Kubernetes's native interfaces.


## 10) Declare underlying resources created by Custom Resource as Annotation on CRD registration YAML

Use an annotation on the Custom Resource Definition to specify the underlying Kubernetes resources that will be created and managed by the Custom Resource. An example of this can be seen for our sample Postgres Custom Resource Definition below:

```
  kind: CustomResourceDefinition
  metadata:
    name: postgreses.postgrescontroller
    annotations:
      composition: Deployment, Service
```

If this information is not exposed, it will be available only in custom controller code and be hidden from end users 
in case they need it for traceability or any other reason. Externalizing this information 
also makes it possible to build tools like [kubediscovery](https://github.com/cloud-ark/kubediscovery) 
that show Object composition tree for custom resource instances built leveraging this information.


## 11) Make your Custom Resource Type definitions compliant with Kube OpenAPI

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

## 12) Define Custom Resource Spec Validation rules as part of Custom Resource Definition

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
    composition: Deployment, Service
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


# Packaging guidelines

## 13) Generate Kube OpenAPI Spec for your Custom Resources

We have developed a [tool](https://github.com/cloud-ark/kubeplus/tree/master/openapi-spec-generator) that you can use for generating Kube OpenAPI Spec for your custom resources. 
It wraps code available in [kube-openapi repository](https://github.com/kubernetes/kube-openapi) 
in an easy to use script. You can use this tool to generate OpenAPI Spec for your custom resources.
The generated Kube OpenAPI Spec documentation for sample Postgres custom resource is 
[here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/openapispec.json).
The OpenAPI Spec for your Operator provides a single place where documentation is available for the entire Type
definition hierarchy for the custom resources defined by your Operator.


## 14) Package Operator as Helm Chart

Create a Helm chart for your Operator. The chart should include two things:

  * Registration of all Custom Resources managed by the Operator. Examples of this can be seen in 
CloudARK [sample Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml) and in 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).

  * Kube OpenAPI Spec for your custom resources. The Spec will be useful as reference documentation and can be leveraged by different tools.


# Documentation guidelines

## 15) Document how your Operator uses namespaces

For Operator developers it is critical to consider how their Operator works with/uses namespaces. There are at least four options:

  * Operator runs in default namespace and Custom Resources are created in default namespace.

  * Operator runs in default namespace but Custom Resources can be created in non-default namespaces.

  * Operator runs in a non-default namespace and Custom Resources are created in default namepsace.

  * Operator runs in a non-default namespace and Custom Resources can be created in non-default namespaces.

Given these options, it will help consumers of your Operator if there is a clear documentation of how namespaces 
are used by your Operator.


## 16) Document Service Account usage of your Operator

Your Operator may be using default service account or some specific service account. Moreover, the service account
may need to be granted specific permissions. Clearly document the service account needs of your Operator.


## 17) Document naming convention and labels needed to be used with Custom Resources

You may have special requirements for naming your custom resource instances or some of their
Spec properties. Similarly you may have requirements related to the labels that need to be added on them.
Document this information with OpenAPI Spec annotations that you will define for your Type definitions. 
That way this information will help application developers when they are trying to compose/use your custom resources with custom resources from other Operators.


## Evaluation of community Operators

Here is a table showing conformance of different community Operators to above guidelines.

| Operator      | URL           | Guidelines satisfied  | Comments     |
| ------------- |:-------------:| ---------------------:| ------------:|
| Oracle MySQL Operator | https://github.com/oracle/mysql-operator | 2, 3, 4, 5, 6, 7 | 1: Not satisfied because of exposing mysqldump in Spec <br> 8: Not satisfied as composition of CRDs not defined <br>9, 10: Not satisfied, PR that addresses them: https://github.com/oracle/mysql-operator/pull/216 
| PressLabs MySQL Operator | https://github.com/presslabs/mysql-operator  | 1, 2, 3, 5, 6, 7, 9, 10, 11 | 4: Not satisfied because CRD installed in Code <br> 8: Not satisfied as composition of  CRDs not defined
| CloudARK sample Postgres Operator | https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd-v2 | 1, 2, 3, 4, 5, 6, 8, 9, 10, 11 | 



If you are interested in getting your Operator checked against these guidelines, 
[create a Issue](https://github.com/cloud-ark/kubeplus/issues) with link to your Operator Code and we will analyze it.









