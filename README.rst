=========
KubePlus
=========

Kubernetes provides ability to extend a cluster's functionality by adding Custom Resource Controllers.

Such a extended Kubernetes cluster essentially represents a purpose-built application platform
which may consist of several custom resource controllers.

*Discoverability* is a key requirement for such platforms.

Kubernetes admins and application developers should be able to self-discover the
capabilities of such purpose built application platform.

Here we provide generic guidelines for developing Kubernetes Custom Resource Controllers
to aid discoverability. Any custom resource controller that is written following
these guidelines will make it easy for Kubernetes admins and application developers to discover
the capabilities of their purpose built application platform on Kubernetes.


**Explicitly identify resource life-cycle actions in Custom Resource Type definition**:

When definining the type for a custom resource you should explicitly document
and identify the various life-cycle actions that can be performed on a resource by the
custom resource controller. For instance, a custom resource controller for
Postgres might identify following life-cycle actions that it can perform
on a custom Postgres resource - create_database, add_user, modify_user_password,
modify_user_permissions, etc. The actions can be defined in the resource spec
as method signatures or can be defined as achievable states of the resource.

By explicitly identifying such actions in the type definition it will be possible
to write discoverability tools that inspect the type definition and discover
all the actions that can be done on a custom resource by that controller.


**Expose resource's configuration parameters as Controller ConfigMaps**:

A controller should be written such that it takes inputs for underlying resource's
configuration parameters through ConfigMap(s). For instance, a custom resource controller
for Postgres should identify the parameters from postgresql.conf
that it will support through resource-specific ConfigMaps.
The controller should be written to generate the required configuration file(s) from the
values defined in such ConfigMaps.

This approach of elevating the underlying resource's configuration parameters as ConfigMaps,
makes them first class entities in Kubernetes. This allows admins and application developers to 
change a resource's settings in Kubernetes-native manner, without having to worry about how to update appropriate
configuration file(s) and how to then inject them in appropriate Pods.

Moreover, these parameters should be explicitly identified in the resource type definition.
This will enable discoverability tool(s) to inspect the type definition and determine
the configurable settings of the underlying resource.


**Use annotations to pass parameters that modify the behavior of the Custom Resource Controller itself**:

Use Kubernetes annotations to customize the behavior of the custom resource controller itself.
Such annotations should be defined on the resource that is used to register the custom resource type ('kind').

The corresponding custom resource controller should be written to inspect these annotations and 
configure itself based on their values. For instance, a Postgres custom resource controller might have
a parameter that defines the timeout value for checking if a Postgres Pod has become
active or not. This timeout value should be specified as a annotation on that CRD's
registration resource.

Using annotations instead of ConfigMap allows discovery tools to inspect
a single place to understand the configuration options for a controller.


**Explicitly identify underlying resources that will be created by the Custom Resource Controller**:

A controller will typically create one or more Kubernetes resources, such as Pod, Service, Deployment, Secret, Ingress, etc.,
as part of handling a custom resource. From understandability point-of-view,
it is important to know what are all such resources that will be created by a custom resource controller.

We recommend that the type definition be used to identify all such resources that will be created by a custom resource controller.
This will allow discovery and provenance tools to generate a complete picture of the runtime
behavior of a custom resource controller enabling 'what-if' analyses. This can be especially important when a cluster is
shared by multiple teams.















