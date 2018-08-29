# Operator Guidelines

Kubernetes provides ability to extend a cluster's functionality by adding new Operators (Custom
Resource Definitions + custom controllers).

We have come up with following best practice guidelines for developing Kubernetes Operators with the goal to bring uniformity across multiple Operators. These would improve usability of Operators and enable users to consume them in a group to build custom application platforms. 

These guidelines are based on our study of various Operators
written by the community and through our experience of building
[discovery](https://github.com/cloud-ark/kubediscovery) and [provenance](https://github.com/cloud-ark/kubeprovenance) 
tools for Kubernetes Operators.


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


## 3) Use kube-openapi annotations in Custom Resource Type definition

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

One approach towards this is to write your Custom controller to collect required metrics.
An example of this can be seen in the [MySQL Operator](https://github.com/oracle/mysql-operator/blob/master/docs/setup/monitoring.md).
Another approach is to use a generic tool such as [kubeprovenance](https://github.com/cloud-ark/kubeprovenance) in your cluster.
KubeProvenance uses Kubernetes Audit Logs to build lineage for custom resources.
Additionally, it provides various provenance query operators to query the collected custom resource provenance information.



## Evaluation with example Operators

Here is a table showing conformance of different community Operators to above guidelines.

| Operator      | URL           | Guidelines satisfied  | Comments     |
| ------------- |:-------------:| ---------------------:| ------------:|
| Oracle MySQL  | https://github.com/oracle/mysql-operator | 2, 4, 5, 6, 8 | 1: Not satisfied because of exposing mysqldump in Spec
| PressLabs MySQL| https://github.com/presslabs/mysql-operator  | 1, 2, 3, 5, 6 | 4: Not satisfied because CRD installed in Code
| CloudARK Postgres| https://github.com/cloud-ark/kubeplus/tree/master/postgres-crd-v2 | 1, 2, 3, 4, 7 | 5, 6: Work-in-Progress



If you are interested in getting your Operator checked against these guidelines, 
[create a Issue](https://github.com/cloud-ark/kubeplus/issues) with link to your Operator Code and we will analyze it.









