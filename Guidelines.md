# KubePlus

Kubernetes provides ability to extend a cluster's functionality by adding new operators (Custom
Resource Definitions + associated controllers). Such an extended Kubernetes cluster essentially 
represents a purpose-built application platform.

*Discoverability* is a key requirement for these custom operators.

Here we provide generic guidelines for developing Kubernetes Operators that will help
with this requirement. Any custom operator that is written following
these guidelines will make it easy for Kubernetes admins and application developers to self-discover
its capabilities. This will also bring consistency in using multiple such operators
in a single Kubernetes cluster to form a purpose-built application platform.


## 1) Define the desired states of a Custom Resource as declarative Spec in its Type definition 

Custom Resource Type definitions should use declarative format for representing the state of the Custom Resource.
Users should be able to specify the desired state of the custom resource using yaml input to kubectl only.
Kubernetes APIs are designed such that the desired state of an object is sent to the API server, 
and the cluster works to reconcile the actual state with the desired state. 
So Custom Resource Type definitions should declare a canonical data structure to capture the state of the object it is defining. 
Life-cycle actions should be embedded in the controller logic when it reconciles actual state to desired state. 

For example, to add a new user to a Postgres custom resource, 
users should just update the yaml definition of Postgres resource instance adding a 
[new name in the users list](https://github.com/cloud-ark/kubeplus/blob/master/postgres-crd-v2/artifacts/examples/initializeclient.yaml). 
The controller code should be written to perform diff of the current users with the desired user 
and perform the required actions (such as add the new user) based on the received desired state. 

Explicit documentation of the supported life-cycle actions by the operator and how 
those can be performed using its declarative definition is recommended for the ease of use.


## 2) Custom Resource Controller implementation should be state/level based and not trigger/edge based

Custom controller code should be written such that it reconciles the current state
with the desired state as defined in the Custom resource's Spec 
(i.e. it is [level-based](https://stackoverflow.com/questions/31041766/what-does-edge-based-and-level-based-mean)). 

It should *not* matter *when* the desired state change is requested. 
This makes the controller code more robust as it does not have to worry about missing a state change trigger. 


## 3) Expose Custom Resource's configuration parameters as ConfigMaps or Annotations

A controller should be written such that it takes inputs for underlying resource's
configuration parameters through ConfigMap(s) or Annotations. For instance, a custom resource controller
for Postgres should identify the parameters from postgresql.conf
that it will support through resource-specific ConfigMaps.
The controller should be written to generate the required configuration file(s) from the
values defined in such ConfigMaps/Annotations and propagate 

This approach of elevating the underlying resource's configuration parameters
makes them first class entities in Kubernetes. This allows admins and application developers to 
change a resource's settings in Kubernetes-native manner, without having to worry about how to update appropriate
configuration file(s) and how to then inject them in appropriate Pods.

Such configuration parameters should be explicitly identified and documented for each custom resource.


## 4) Explicitly identify underlying resources that will be created by the Custom Resource Controller

A custom controller will typically create one or more Kubernetes resources, such as Pod, Service, Deployment, Secret, Ingress, etc.,
as part of instantiation of its custom resource. It is important for a user to understand this underlying composition
of a custom resource. We recommend clear documentation of this composition with chosen nomenclature.
We also recommend adding labels to each of these underlying objects with custom resource instance UUID 
to maintain associations.


## 5) Use annotations to pass parameters that modify the behavior of the Custom Resource Controller itself

Use Kubernetes annotations to customize the behavior of the custom resource controller itself.
Such annotations should be defined on the resource that is used to register the custom resource type ('kind').

The corresponding custom resource controller should be written to inspect these annotations and 
configure itself based on their values. For instance, a Postgres custom resource controller might have
a parameter that defines the timeout value for checking if a Postgres Pod has become
active or not. This default timeout value can be modified as an annotation on that CRD's
registration resource.