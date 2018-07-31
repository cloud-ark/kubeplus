=========
KubePlus
=========

Purpose-built application platforms on Kubernetes.

KubePlus Purpose-built Platforms extend Kubernetes with Operators of your choice.
This allows embedding customer-specific workflows and platform life-cycle actions directly in Kubernetes.
Common examples of such Kubernetes Operators are for Platform elements such as 
Postgres, MySQL, Fluentd, Prometheus etc.

**Value of KubePlus**

*1) Choose your own platform elements*

KubePlus enables you to Build Your Own Platform on Kubernetes. You can choose your platform elements for databases, caching, logging, monitoring etc. 
KubePlus extends your Kubernetes cluster with Kubernetes Operators for those specific platform elements.
Examples of such operators can be MySQL, Ngnix, Redis etc. 


*2) Eliminate out-of-band platform automation*

Kubernetes Operators embed platform element life-cycle actions directly in Kubernetes. An example of a Kubernetes Operator can be Postgres Operator that 
embeds life-cycle actions such as create a database, add user to the database, change password of a user etc.
Such Operators leverage Kubernetes's strength of control loop (current state -> desired state) eliminating additional out-of-band automation.


*3) Consistency across Kubernetes Operators*

Based on our study of existing Kubernetes Operators, we have come up with common guidelines that need to be followed by any Operator to be part of KubePlus. 
This brings consistency and quality in packaging Kubernetes Operators to build a purpose-built platform.


*4) Improved usability of Kubernetes Operators*

KubePlus installs an additional components, KubePlus Discovery Manager, on your Kubernetes cluster to improve usability of custom Operators.

It provides following information about newly added custom resources:

- Static information like Life-cycle actions that can be performed on a custom resource (e.g. You can add/remove users to an instance of MySQL resource.)

- Dynamic information like Composition of custom resources in terms on native Kubernetes resources (e.g. If you create an instance of a MySQL custom resource, it would internally create a pod and a service.)

- Custom resource specific configurable parameters exposed by the controller (e.g. MySQL configurable parameters)


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

KubePlus will install two additional component: KubePlus Operator Manager and KubePlus Discovery Manager. KubePlus Operator Manager enables installation and lifecycle management of required Operators. KubePlus Discovery Manager improves consumability of bundled Operators by providing additional information about them. 

KubePlus does not introduce any new CLI interface. 
Entire workflow is supported through native Kubernetes interface of kubectl. 

*Purpose-built platform deployment*

Once core KubePlus components (Operator Manager and Discovery Manager) are installed on the cluster, Kubernetes cluster administrator defines Kubernetes Operators to be installed in an extensions.yaml file and then uses following command to install those Operators. 

`$ kubectl apply -f extensions.yaml`

*Purpose-built platform usage*

Kubernetes app developers can create/delete/update/list the newly added 
custom resources by directly using kubectl CLI. e.g. 

`$ kubectl apply -f postgres.yaml`

Additionally they can get more information about the composition and life-cycle actions of these resources through KubePlus Discovery Manager. e.g. 

`$ kubectl get –raw …/postgres/*/composition`

Composition of postgres in terms of underlying Kubernetes resources like Pod, Service etc. 

`$ kubectl get –raw …/postgres/explain`

Postgres life cycle actions and other static information



**Available Operators**

1) Postgres
   - Check postgres-crd-v2/README.rst for details about how to use this Operator.


**Build your own Operators**

If you are interested in building your own operators, check the steps that you can follow here_:

.. _here: https://github.com/cloud-ark/kubeplus/issues/14

