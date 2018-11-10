=========
KubePlus
=========

KubePlus delivers `Platform as Code`__ experience on Kubernetes.

.. _pac: https://medium.com/@cloudark/evolution-of-paases-to-platform-as-code-in-kubernetes-world-74464b0013ca

__ pac_


KubePlus Platform Kit
======================

`Kubernetes Operators`__ extend Kubernetes API to manage
third-party software as native Kubernetes objects. Today, number of Operators are
being built for middlewares like databases, queues, loggers, etc. This has led to
tremendous choice in the platform elements for building application platforms.
Current popular approach is to ‘self-assemble’ platform stacks using Kubernetes Operators of
choice. This requires significant efforts and there is 
lack of consistent user experience across multiple Operators.

.. _Operators: https://medium.com/@cloudark/why-to-write-kubernetes-operators-9b1e32a24814

__ Operators_


KubePlus Platform Kit streamlines the process of composing multiple Operators into a custom PaaS,
and enables creating application Platforms as Code. Using KubePlus Platform Kit,

* DevOps Engineer constructs a custom PaaS comprised of required Kubernetes Operators.

* Application Developer declares and creates application platforms as code leveraging custom resources
  introduced by the installed Operators.

We bring consistency of usage across multiple Operators with our Operator development guidelines_.
Teams can Build their Own PaaSes on Kubernetes selecting required Operators 
from our `repository of certified Operators`__ packaged as Helm charts.

.. _guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md

.. _repository: https://github.com/cloud-ark/operatorcharts/blob/master/index.yaml

__ repository_


.. image:: ./docs/KubePlus-Flow.jpg
   :scale: 75%
   :align: center


Demo
====

https://drive.google.com/file/d/1jDptIWM8fiAorlZdW-pwOMttxAQAZHIR/view


Usage
======

KubePlus is designed with 3 user personas in mind. 

*1. Operator developer*

*2. Cluster Administrator*

*3. Application developer*

KubePlus does not introduce any new CLI. KubePlus users continue to use the
standard Kubernetes CLI (kubectl) and YAML definition format to manage their platforms.


 
.. image:: ./docs/KubePlus-Platform-Kit.jpg
   :scale: 75%
   :align: center


Value of KubePlus
==================

*1) Uniformity between native and custom Kubernetes resources*

Our Operator development guidelines are designed to ensure custom resources become 
first-class entities of Kubernetes. 

*2) No new CLI to learn*

KubePlus does not introduce any new CLI. Users can work with the same Kubernetes native interfaces like kubectl and YAML to leverage KubePlus functionality.


*3) Eliminate out-of-band platform automation*

Kubernetes Operators embed platform element life-cycle actions directly in Kubernetes. An example of a Kubernetes Operator can be Postgres Operator that 
embeds life-cycle actions such as create a database, add user to the database, change password of a user etc.
Such Operators leverage Kubernetes's strength of control loop (current state -> desired state) eliminating additional out-of-band automation.


*4) Common language between Devs and Ops*

KubePlus leverages kubectl for management of Operators by Ops and their consumption by Devs. This makes Kubernetes YAMLs as the common language between Devs and Ops. 


*5) Discovery of custom resources*

KubePlus installs an additional component, KubePlus Discovery Manager, on your Kubernetes cluster to improve usability of custom Operators.

KubePlus Discovery Manager provides information about custom resources managed by the Operators. E.g. Assume there is a Postgres Operator which is managing a custom resource called Postgres. To make it is easy to consume Postgres resource in your application YAML, KubePlus will provide following information about Postgres resource: 

- Static information like OpenAPI Spec for the Postgres resource. This information can be used by application developers when creating their application platform.

- Dynamic information like composition tree of custom resources in terms on native Kubernetes resources (e.g. which Deployment, Pod, Service, etc. objects are part of the composition tree of a Postgres resource instance.)


KubePlus Architecture
======================

.. image:: ./docs/Architecture.jpg
   :scale: 75%
   :align: center


Try it
=======

Minikube
---------

Detailed steps are available in `kubeplus-steps.txt`__.

.. _steps: https://github.com/cloud-ark/kubeplus/blob/master/kubeplus-steps.txt

__ steps_


Here is a summary:


**1) Install KubePlus (by cluster administrator)**

KubePlus requires Helm to be installed on the cluster.

Install Helm:

  ``$ helm init``

Once tiller pod is Running (kubectl get pods -n kube-system), install KubePlus.
We provide deployment YAMLs for deploying KubePlus.


  ``$ kubectl apply -f deploy/``

Check KubePlus is ready

  ``$ kubectl get pods``

KubePlus consists of 4 containers - operator-manager, operator-deployer, kube-discovery-apiserver, etcd.
Wait till all 4 containers come up and are in 'Running' state (4/4 READY).

**2) Create custom PaaS (by cluster administrator)**


a) Once core KubePlus is READY, Kubernetes cluster administrators define Kubernetes Operators to be installed in yaml files (e.g.: Postgres_, MySQL_, Moodle_) 
and use following kubectl commands:

.. _Postgres: https://github.com/cloud-ark/kubeplus/blob/master/postgres-operator.yaml

.. _MySQL: https://github.com/cloud-ark/kubeplus/blob/master/mysql-operator-chart-0.2.1.yaml

.. _Moodle: https://github.com/cloud-ark/kubeplus/blob/master/moodle-operator.yaml


b) Deploy/install Operators:

  ``$ kubectl apply -f <operator yaml file>``


c) Find out all the installed Operators:

  ``$ kubectl get operators``



**3) Create Application Platform as Code (by application developer)**

Kubernetes application developers can create/delete/update/list the newly added 
custom resources by using kubectl CLI using following commands:

a) Find out custom resource Kinds managed by an Operator:

  ``$ kubectl describe operators postgres-operator``

  ``$ kubectl describe customresourcedefinition postgreses.postgrescontroller.kubeplus``

b) Find out details about a custom Kind:

  ``$ kubectl get --raw "/apis/kubeplus.cloudark.io/v1/explain?kind=Postgres"  | python -m json.tool``

c) Define application Platform elements_:

  ``$ vi platform.yaml``

.. _elements: https://github.com/cloud-ark/kubeplus/blob/master/platform.yaml


d) Create application Platform:

  ``$ kubectl apply -f platform.yaml``

e) Find out dynamic composition tree for Postgres custom resource instance:

  ``$ kubectl get --raw "/apis/kubeplus.cloudark.io/v1/composition?kind=Postgres&instance=postgres1" | python -mjson.tool``


Deploy Multiple Operators to create a custom PaaS
-------------------------------------------------

a) Install Helm and KubePlus like above

b) Deploy multiple Operators through single YAML file

   ``$ kubectl create -f paas.yaml``

c) Check deployed operators

   ``$ kubectl get operators``

d) Describe Operators

   ``$ kubectl describe operators postgres-operator``

   ``$ kubectl describe operators moodle-operator``

   ``$ kubectl describe operators mysql-operator-0.2.1``

e) Find out custom resource Kinds registered by Operators

    ``$ kubectl describe customresourcedefinition postgreses.postgrescontroller.kubeplus``

    ``$ kubectl describe customresourcedefinition moodles.moodlecontroller.kubeplus``

f) Explain custom Kinds

   ``kubectl get --raw "/apis/kubeplus.cloudark.io/v1/explain?kind=Postgres"  | python -m json.tool``
   
   ``kubectl get --raw "/apis/kubeplus.cloudark.io/v1/explain?kind=Postgres.PostgresSpec"  | python -m json.tool``

   ``kubectl get --raw "/apis/kubeplus.cloudark.io/v1/explain?kind=Postgres.PostgresSpec.UserSpec"  | python -m json.tool``

   ``kubectl get --raw "/apis/kubeplus.cloudark.io/v1/explain?kind=Moodle"  | python -m json.tool``


Real cluster:
--------------

- Moodle

  - Deploy Moodle Operator and then create Moodle Instance on a EC2 instance

    - Follow examples/moodle/steps.txt


Operator Development Guidelines
================================

Checkout_ our guidelines for developing Operators.
These guidelines are based on our study of various Operators written by the community
and through our experience of building Operators ourselves along with discovery_ and provenance_ tools for Kubernetes.

.. _Checkout: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md

.. _discovery: https://github.com/cloud-ark/kubediscovery

.. _provenance: https://github.com/cloud-ark/kubeprovenance


Available Operators
--------------------

https://github.com/cloud-ark/operatorcharts


If you are interested in building your own operators, you can follow steps here_.

.. _here: https://github.com/cloud-ark/kubeplus/issues/14

You can also use tools like kubebuilder_ or `Operator SDK`__ to build your Operator.

.. _kubebuilder: https://github.com/kubernetes-sigs/kubebuilder

.. _sdk: https://github.com/operator-framework/operator-sdk

__ sdk_


Issues
======

Suggestions/Issues are welcome_.

.. _welcome: https://github.com/cloud-ark/kubeplus/issues

