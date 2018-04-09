=========
KubePlus
=========

Purpose-built application platforms on Kubernetes.

KubePlus Purpose-built Platforms extend Kubernetes with custom resources.
This allows embedding platform lifecycle actions and customer-specific workflows directly in Kubernetes.
Common examples of such extensions are Postgres, MySQL, Fluentd, Prometheus etc.

**Value of KubePlus**

*1) Choose your own platform elements*

KubePlus allows you to build your own application platform on Kubernetes. You can choose your platform elements for databases, caching, logging, monitoring etc. And then KubePlus extends your Kubernetes cluster adding those platform elements as Kubernetes custom resources.

*2) Eliminate out-of-band platform automation*

KubePlus allows embedding platform element lifecycle actions or workflows in Kubernetes custom resources. Examples of such lifecycle actions can be creation of a database, add user to the database, change password of a user etc. This leverages Kubernetes strength of control loop (current state -> desired state) and eliminates additional out-of-band automation.


*3) Consistency across Kubernetes custom resources*

Based on our study of existing Kubernetes custom controllers/extensions, we have come up with common guidelines that need to be followed by any custom controller to be part of KubePlus. This brings consistency and quality in packaging Kubernetes extensions to build a purpose-built platform. 


*4) Improved usability of Kubernetes custom resources*

KubePlus installs an additional software component, named KubeARK, on your Kubernetes cluster to improve usablity of new custom resources. 

KubeARK provides following information:

- Resource specific configurable parameters exposed by the controller (e.g. MySQL configurable parameters)

- Lifecycle actions that can be performed on a custom resource   (e.g. You can add/remove users to an instance of MySQL resource.)

- Composition of custom resources in terms on native Kubernetes resources (e.g. If you create an instance of a MySQL custom resource, it internally would create a pod and a service.)


**How it Works?**

Imagine an EdTech startup building a classroom collaboration application on Kubernetes.
They have following high level requirements for their application platform:

- Platform should be composable for adding / updating platform elements.

- Platform learning curve for developers should be minimal without having to learn any new tools/interfaces for creating/managing the application platform. 

This application requires following platform elements with the ability perform their lifecycle operations.

- Nginx: (Sample lifecycle operations- Add/Remove endpoints, Set routes, Set SSL etc.)

- Postgres: (Sample lifecycle operations- Create db, Add/remove users and passwords.)

- Backup: (Sample lifecycle operations- Backup Postgres db, Restore Postgres db.)

- Prometheus: (Sample lifecycle operations- Define monitoring endpoints. Set metrics.)

- Fluentd: (Sample lifecycle operation- Set log rotation policy.)

KubePlus purpose built platform for this EdTech startup would contain 5 custom resources - Nginx, Postgres, Backup, Prometheus and Fluentd, which are written to follow our guidelines for Kubernetes extensions. 

KubePlus will also install a complimentary extension called KubeARK to improve consumability of the custom resources. KubeARK provides additional information about KubePlus custom resources using kubectl interface. 

Kubernetes admin deploys KubePlus on a Kubernetes cluster using following simple commands.

- cld platform create platform.yaml -- plaform.yaml defines the custom resources to be added to the kubernetes cluster.

- cld platform update platform.yaml -- Add/Update custom resources.

- cld platform list -- List installed custom resources.

Kubernetes users / app developers can create/delete/update/list the newly added 5 resources by directly using kubectl CLI

- kubectl create -f postgres.yaml

Additionally they can use KubeARK to get more information about composition, configurables and supported operations on these resources.

-  kubectl explain kubeark.postgres

This command provides following more information about postgres:

- Composition - Native kubernetes resources (like pods, service, deployment etc) that will be created when postgres custom resource is created, 

- Configurables - postgres configurable parameters that can be configured

- Supported operations - Lifecycle operations that can be performed e.g. add/remove users.


**Example CRDs**

1) Postgres
   - Check postgres-crd/README.rst for details of this CRD
