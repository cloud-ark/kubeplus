=========
KubePlus
=========

Purpose-built application platforms on Kubernetes.

KubePlus Purpose-built Platforms extend Kubernetes Operators of your choice.
This allows embedding customer-specific workflows and platform life-cycle actions directly in Kubernetes.
Common examples of such Kubernetes Operators can be Postgres, MySQL, Fluentd, Prometheus etc.

**Value of KubePlus**

*1) Choose your own platform elements*

KubePlus enables you to Build Your Own Platform on Kubernetes. You can choose your platform elements for databases, caching, logging, monitoring etc. 
KubePlus extends your Kubernetes cluster with Kubernetes Operators for those specific platform elements.

*2) Eliminate out-of-band platform automation*

Kubernetes Operators embed platform element life-cycle actions directly in Kubernetes. An example of a Kubernetes Operator can be Postgres Operator that 
embeds life-cycle actions such as create a database, add user to the database, change password of a user etc.
Such Operators leverage Kubernetes's strength of control loop (current state -> desired state) eliminating additional out-of-band automation.

*3) Consistency across Kubernetes custom resources*

Based on our study of existing Kubernetes Operators, we have come up with common guidelines that need to be followed by any Operator to be part of KubePlus. 
This brings consistency and quality in packaging Kubernetes Operators to build a purpose-built platform.


*4) Improved usability of Kubernetes custom resources*

KubePlus installs an additional component, named CRDProvenanceAPIServer, on your Kubernetes cluster to improve usability of custom Operators.

CRDProvenanceAPIServer provides following information:

- Custom resource specific configurable parameters exposed by the controller (e.g. MySQL configurable parameters)

- Life-cycle actions that can be performed on a custom resource (e.g. You can add/remove users to an instance of MySQL resource.)

- Composition of custom resources in terms on native Kubernetes resources (e.g. If you create an instance of a MySQL custom resource, it would internally create a pod and a service.)


**How it Works?**

Imagine an EdTech startup building a classroom collaboration application on Kubernetes. They have following high level requirements for their application platform:
- Platform should be composable. It should be possible to add or update required platform elements to it.
- Platform learning curve for developers should be minimal.

This application requires following platform elements.

- Nginx for load balancing: (Required life-cycle actions- Add/Remove routes, Configure SSL Certificates.)

- Postgres for backend storage: (Required life-cycle actions- Create/drop db, Backup/restore db, Add/remove users.)

- Prometheus for monitoring: (Required life-cycle actions- Define monitoring endpoints, Set metrics.)

- Fluentd for logging: (Required life-cycle action- Set log rotation policy.)


**KubePlus Purpose-built Platform**

KubePlus purpose-built platform for this EdTech startup would contain four custom operators - Nginx, Postgres, Prometheus and Fluentd, which are written to 
follow our guidelines for Kubernetes Operators.

KubePlus will install a special Kubernetes API Server named CRDProvenanceAPIServer to improve consumability of these KubePlus Operators.
CRDProvenanceAPIServer provides additional information about KubePlus custom resources using kubectl interface.

*KubePlus deployment*

Kubernetes admin deploys KubePlus on a Kubernetes cluster using following simple commands.

- *kubeplus create platform platform.yaml*: platform.yaml defines the custom operators to be added to the Kubernetes cluster.
  It requires following information about each custom operator that needs to be part of KubePlus platform.

  - Path to operator's helm chart.

  - KubePlus specific yaml file per operator that provides additional information about configurables, life-cycle actions, and composition. 

- *kubeplus update platform platform.yaml*: Add/Update custom operators.

- *kubeplus list platform*: List installed custom operators.


*KubePlus usage*

Kubernetes users can create/delete/update/list the newly added custom resources by directly using kubectl CLI. e.g. # kubectl apply -f postgres.yaml
Additionally they can use CRDProvenanceAPIServer to get more information about the composition, configurables, and life-cycle actions of these resources. e.g.

- *kubeplus get configurables Postgres*: This command provides information about supported configurable parameters of Postgres custom resource.

- *kubeplus get composition Postgres*: This command provides information about composition of Postgres custom resource
  in terms of underlying Kubernetes resources like Pod, Service etc.

- *kubeplus get composition Postgres Postgres_wordpress*: Here Postgres_wordpress is an instance of Postgres custom resource.
  This command shows actual underlying Kubernetes resources created for Postgres_wordpress resource instance.

- *kubeplus get actions Postgres*: This command provides information about supported life-cycle actions on Postgres resource
  like backup/restore db and how they can be performed using declarative yaml definition.


**Available Operators**

1) Postgres
   - Check postgres-crd-v2/README.rst for details about how to use this Operator.


**Build your own Operators**

If you are interested in building your own operators, check the steps that you can follow here_:

.. _here: https://github.com/cloud-ark/kubeplus/issues/14

