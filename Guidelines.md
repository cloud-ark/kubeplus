# Operator Guidelines

Kubernetes provides ability to extend a cluster's functionality by adding new Operators (Custom
Resource Definitions + associated controllers). 

We have come up with following best practice guidelines for developing Kubernetes Operators.
Operators developed following these guidelines provide ease of use and management.

These guidelines are based on our study of various Operators
written by the community and through our experience of building
[discovery](https://github.com/cloud-ark/kubediscovery) and [provenance](https://github.com/cloud-ark/kubeprovenance) 
tools for Kubernetes Operators.


## 1) Prefer declarative state over imperative actions in Custom Resource Spec definition

Define the desired states of a Custom Resource as declarative Spec in its Type definition.
Any updates to the state of a Custom Resource instance should be defined solely as the desired
state in the declarative Spec of the resource.
Users should not be concerned with the procedural details of specifying changes from the previous state.
For example, to add a new user to a Postgres custom resource, 
users should just update the yaml definition of Postgres resource instance adding a 
[new name in the users list](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/artifacts/examples/add-user.yaml).

A consequence of this is that Custom controller code should be written such that it reconciles the current state
with the desired state by performing diff of the current state with the desired state. 
Life-cycle actions of the underlying resource should be embedded in the controller logic.
For example, Postgres custom resource controller should be written to perform diff of the current users with the desired user
and perform the required actions (such as adding new users, deleting current users, etc.) based on the received desired state.

Note that the principle of diff-based implementation for custom controllers is essentially an extension of
the level-triggered approach recommended in the [general guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md) 
for developing Kubernetes controllers.

An example where underlying imperative actions are exposed in the Spec is this 
[MySQL Backup Custom Resource Spec](https://github.com/oracle/mysql-operator/blob/master/examples/backup/backup.yaml#L7).
Here the fact that MySQL Backup is done using mysqldump tool is exposed in the Spec.
In our view such internal details should not be exposed in the Spec as it prevents Custom Resource Type definition 
(in this case, Backup) to evolve without affecting its users.


## 2) Use OwnerReferences with Custom Resource instances

A custom controller will typically create one or more Kubernetes resources, such as Pod, Service, Deployment, Secret, Ingress, etc., 
as part of instantiation of its custom resource. The controller should be written to set OwnerReference on custom
or native Kubernetes Kinds that it would create. These references aid with supporting
[discovery of information](https://github.com/cloud-ark/kubediscovery), such as the Object composition tree, for custom resource instances
and garbage collection of resources.

Examples of Operators that use OwnerReferences include, [Etcd Operator](https://github.com/coreos/etcd-operator/blob/master/pkg/cluster/cluster.go#L351),
[Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/controller.go#L508), and 
[MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/pkg/resources/services/service.go#L34).


## 3) Use kube-openapi annotations in Custom Resource Type definition

When defining the types corresponding to your custom resources, you should use
kube-openapi annotation - ``+k8s:openapi-gen=true''
in the type definition to [enable generating documentation for the custom resource](https://medium.com/@cloudark/understanding-kubectl-explain-9d703396cc8).
An example of this annotation on type definition is our [Postgres operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/pkg/apis/postgrescontroller/v1/types.go#L28). This annotation enables generating OpenAPI Spec documentation for custom resources as seen [here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/openapispec.json). We have developed a [tool](https://github.com/cloud-ark/kubeplus/tree/master/openapi-spec-generator) 
that you can use for generating OpenAPI Spec for your custom resources. 


## 4) Package Operator as Helm Chart and register Custom Resources as part of it

You should create a Helm chart for your Operator. The chart should include two things: 

(1) Registration of all Custom Resources managed by the Operator.
Examples of this can be seen in our [Postgres Operator](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/postgres-crd-v2-chart/templates/deployment.yaml)
and in this [MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/mysql-operator/templates/01-resources.yaml).
Registering CRDs as part of Helm Chart, instead in [Golang Code](https://github.com/coreos/etcd-operator/blob/master/pkg/controller/backup-operator/operator.go#L76),
has following advantages: (a) Helm Charts have become the standard mechanism for defining all installation artifacts.
(b) You can use this approach even if your Operator is not written in GO.

(2) Any documentation for the custom resources, such as the OpenAPI Spec, for your custom resources.
The documentation of the custom resources will be useful for application developers to figure out how to use your custom resources.



## 5) Support Operator configuration parameters

Custom controller of an Operator will need to support some form of customization. For example, 
[this MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/tutorial.md#configuration) supports following customization settings: whether to deploy
the Operator cluster-wide or within a particular namespace, which version of MySQL should be installed, etc.
If you have followed previous guideline and have Helm chart for your Operator then use Helm's values YAML file to specify
such parameters. If not, use Operator-specific ConfigMap for this purpose. 



## 6) Support control over underlying resource's configuration parameters

A controller should be written such that it takes inputs for underlying resource's
configuration parameters. We have seen three different approaches towards this so far based on using ConfigMaps, using Annotations, using Spec definition itself.
[Nginx Custom Controller](https://github.com/nginxinc/kubernetes-ingress/tree/master/examples/customization) supports both ConfigMap and Annotation.
[Oracle MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/user/clusters.md) uses ConfigMap.
[PressLabs MySQL Operator](https://github.com/presslabs/mysql-operator) uses Custom Resource [Spec definition](https://github.com/presslabs/mysql-operator/blob/master/examples/example-cluster.yaml#L22).



## 7) Define composition of Custom Resources using annotation on the Custom Resource YAML Definition

We recommend that you use an annotation on the Custom Resource Definiton to identify the underlying Kubernetes resources
that will be created by the Custom Resource. An example of this can be seen for our Postgres resource 
[here](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/artifacts/deployment/deployment.yaml#L33).
Doing so will enable tools like kubediscovery to correctly show composition information for custom resource instances.


## 8) Custom Resource Metrics Collection

You should plan for collecting different metrics for custom resources managed by your Operator.
One approach towards this is to write your Custom controller to collect required metrics.
An example of this can be seen in the [MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/setup/monitoring.md).
Another approach is to use a generic tool such as [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) in your cluster.
KubeProvenance uses Kubernetes Audit Logs to build lineage for custom resources.
It also provides various provenance query operators to query the collected information.


We consider first five guidelines (1-5) as must have for any Operator. Guidelines 7 and 8 are nice to have.
Guideline #6 depends on the nature of the Operator. If your Operator is not managing any underlying resource
such as a database, then this guideline does not apply. 
Guideline #7 allows easy discovery of your custom resource instances.
Guideline #8 is useful for understanding effect of 
performing various actions on your custom resources over time.


Here is our analysis of different Operators for their conformance to above guidelines. 
If you are interested in getting your Operator checked against these guidelines, create a Issue with
following information: <Link to your Operator Code - github, bitbucket> and we will analyze your Operator.


| Operator      | URL           | Guidelines satisfied  | Comments     |
| ------------- |:-------------:| ---------------------:| ------------:|
| Oracle MySQL  | https://github.com/oracle/mysql-operator | 2, 4, 5, 7, 8 | 1: Not satisfied because of exposing mysqldump in Spec
| PressLabs MySQL| https://github.com/presslabs/mysql-operator  | 1, 2, 3, 5, 6 | 4: Not satisfied because CRD installed in Code










