# Operator Guidelines

Kubernetes Operators extend Kubernetes API to manage third-party software as native Kubernetes objects. 
Number of Operators are being built for platform elements like databases, queues, loggers, etc. 
We are seeing more and more fit-for-purpose application platforms being created by composing multiple Operators together.

While working on such a custom platform for one of our customers, we observed challenges that arise when using multiple Operators together. 
For example, some of these Operators tend to introduce a new CLI for its end users. 
We reflected more on usability of Operators while building tools that work on multiple Operators 
e.g.: tools for [discovery](https://github.com/cloud-ark/kubediscovery) and [lineage tracking](https://github.com/cloud-ark/kubeprovenance)
of Custom Resources created by Operators. Examples of such usability challenges can be:

  * Some of the Operators introduce new CLIs. Usability becomes an issue when end users have to learn multiple CLIs to use more than one Operators in a cluster.

  * Some of the Operator type definitions do not follow OpenAPI Specification. This makes it hard to generate documentation for custom resources similar to native Kubernetes resources.

Our study of existing community Operators from this perspective led us to come up with Operator development guidelines that will improve overall usability of Operators. The primary goal of these guidelines is : cluster admin should be able to easily compose multiple Operators together to form a platform stack; and application developers should be able to discover and consume Operators effortlessly.



Here are those guidelines:

## 1) Prefer declarative state over imperative actions in Custom Resource Spec definition

Define the desired state of a Custom Resource instance and any updates to it 
as declarative Spec in its Type definition.
Users should not be concerned with the procedural details of specifying changes from the previous state.

Custom controller code should be written such that it reconciles the current state
with the desired state by performing diff of the current state with the desired state. 
Life-cycle actions of the underlying resource should be embedded in the controller logic.
For example, Postgres custom resource controller should be written to perform diff of the current users with the desired user
and perform the required actions (such as adding new users, deleting current users, etc.) based on 
the [received desired state](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/artifacts/examples/add-user.yaml).

Note that the diff-based implementation approach for custom controllers is essentially an extension of
the level-triggered approach recommended in the [general guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md) 
for developing Kubernetes controllers.

An example where underlying imperative actions are exposed in the Spec is this 
[MySQL Backup Custom Resource Spec](https://github.com/oracle/mysql-operator/blob/master/examples/backup/backup.yaml#L7).
Here the fact that MySQL Backup is done using mysqldump tool is exposed in the Spec.
In our view such internal details should not be exposed in the Spec as it prevents Custom Resource Type definition 
to evolve independently without affecting its users.


## 2) Use OwnerReferences with Custom Resource instances

A custom resource instance will typically create one or more Kubernetes resources, such as Deployment, Service, Secret etc., 
as part of instantiating its custom resources. The controller should be written to set OwnerReferences on 
on such native Kubernetes resources that it creates. 
They are key for correct garbage collection of custom resources.
OwnerReferences also help with finding composition tree of your custom resource instances consisting of
native Kubernetes resources (see guideline #7).

Some examples of Operators that use OwnerReferences are: [Etcd Operator](https://github.com/coreos/etcd-operator/blob/master/pkg/cluster/cluster.go#L351),
[Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/controller.go#L508), and 
[MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/pkg/resources/services/service.go#L34).


## 3) Generate OpenAPI Spec for your Custom Resources

When defining the types corresponding to your custom resources, you should use
kube-openapi annotation - ``+k8s:openapi-gen=true''
in the type definition to [enable generating OpenAPI Spec documentation for the custom resource](https://medium.com/@cloudark/understanding-kubectl-explain-9d703396cc8).
An example of this annotation on type definition is our [Postgres operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/pkg/apis/postgrescontroller/v1/types.go#L28).
We have developed a [tool](https://github.com/cloud-ark/kubeplus/tree/master/openapi-spec-generator) 
that you can use for generating OpenAPI Spec for your custom resources. 
The generated OpenAPI Spec documentation for Postgres custom resource is [here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/openapispec.json). 


## 4) Package Operator as Helm Chart and register CRDs as part of it

You should create a Helm chart for your Operator. The chart should include two things: 

* Registration of all Custom Resources managed by the Operator.
Examples of this can be seen in our [Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml)
and in this [MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).
Registering CRDs as part of Helm Chart instead of in [Code](https://github.com/coreos/etcd-operator/blob/master/pkg/controller/backup-operator/operator.go#L76),
has following advantages: 

  * All installation artifacts are available in one place - the Operator's Helm Chart.

  * It is easy to modify and evolve the CRD.

* OpenAPI Spec for your custom resources (if you have followed guideline #3).
The Spec will be useful for application developers to figure out how to use your custom resources.



## 5) Use Helm chart or ConfigMap for Operator configurables

Typically Operators will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have followed previous guideline and have Helm chart for your Operator then use Helm's values YAML file to specify
such parameters. If not, use ConfigMap for this purpose. This guideline ensures that Kubernetes Administrators
can interact and use the Operator using Kubernetes native's interfaces.



## 6) Use ConfigMap or Annotation or Spec definition for Custom Resource configurables

An Operator generally needs to take inputs for underlying resource's configuration parameters. We have seen three different approaches being used towards this in the community and anyone should be fine to use based on your Operator design. They are - using ConfigMaps, using Annotations, or using Spec definition itself. 

[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).

Similar to guideline #5, this guideline ensures that application developers can interact and use Custom Resources using Kubernetes's native interfaces.



## 7) Define composition of a Custom Resource as an Annotation on its YAML Definition

We recommend that you use an annotation on the Custom Resource Definiton to identify the underlying Kubernetes resources
that will be created by the Custom Resource. An example of this can be seen for our Postgres resource 
[here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/artifacts/deployment/deployment.yaml#L33).

By surfacing the composition information as an annotation on CRD, it is possible
to build tools like [kubediscovery](https://github.com/cloud-ark/kubediscovery)
that show Object composition tree for custom resource instances by using the CRD definition.
OwnerReferences (guideline #2) are also crucial in this regard.



## 8) Plan for Custom Resource Metrics Collection

Your Operator design should plan for collecting different metrics for custom resource instances managed by your Operator. This information is useful for understanding effect of performing various actions on your custom resources over time and improves traceability. 

One approach towards this is to use a generic tool such as [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) in your cluster.
KubeProvenance uses Kubernetes Audit Logs to build lineage information for custom resources.
Additionally, it provides various provenance query operators to query the collected custom resource provenance information.
Another approach is to write your Custom controller to collect required metrics.
An example of this can be seen in this [MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/setup/monitoring.md).


## 9) Plan to use kubectl as the primary interaction point

When designing your Operator you should try to support most of its actions through kubectl. 
Kubernetes contains various mechanisms such as Custom Resource Definitions, Aggregated API servers, 
Custom Sub-resources. Refer to [our blog post](https://medium.com/@cloudark/comparing-kubernetes-api-extension-mechanisms-of-custom-resource-definition-and-aggregated-api-64f4ca6d0966) to learn more about them. 
Before considering to introduce new CLI for your Operator, validate if you can use these mechanisms instead.



## Evaluation of community Operators

Here is a table showing conformance of different community Operators to above guidelines.

| Operator      | URL           | Guidelines satisfied  | Comments     |
| ------------- |:-------------:| ---------------------:| ------------:|
| Oracle MySQL Operator | https://github.com/oracle/mysql-operator | 2, 4, 5, 6, 8, 9 | 1: Not satisfied because of exposing mysqldump in Spec <br>3: Not satisfied, PR to address the violation: https://github.com/oracle/mysql-operator/pull/216 <br> 7: Not satisfied as composition of CRDs not defined
| PressLabs MySQL Operator | https://github.com/presslabs/mysql-operator  | 1, 2, 3, 5, 6, 9 | 4: Not satisfied because CRD installed in Code <br> 7: Not satisfied as composition of  CRDs not defined
| CloudARK sample Postgres Operator | https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd-v2 | 1, 2, 3, 4, 7, 9 | 5, 6: Work-in-Progress



If you are interested in getting your Operator checked against these guidelines, 
[create a Issue](https://github.com/cloud-ark/kubeplus/issues) with link to your Operator Code and we will analyze it.









