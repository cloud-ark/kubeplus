=========
KubePlus
=========

KubePlus is a software toolkit that delivers `Platform as Code`__ experience on Kubernetes.
Platform-as-Code offers composable and repeatable way of creating application platforms
leveraging Kubernetes API extensions, also known as Operators. 

.. _pac: https://medium.com/@cloudark/evolution-of-paases-to-platform-as-code-in-kubernetes-world-74464b0013ca

__ pac_


KubePlus Platform Toolkit
==========================

`Kubernetes Operators`__ extend Kubernetes API to manage
third-party software as native Kubernetes objects. Today, number of Operators are
being built for middlewares like databases, queues, loggers, etc. This has led to
tremendous choice in the platform elements for building application platforms.
Current popular approach is to ‘self-assemble’ platform stacks using Kubernetes Operators of
choice. This requires significant efforts and there is 
lack of consistent user experience across multiple Operators.

.. _Operators: https://medium.com/@cloudark/why-to-write-kubernetes-operators-9b1e32a24814

__ Operators_


KubePlus Platform toolkit streamlines the process of composing multiple Operators into a custom PaaS,
and enables creating application Platforms as Code. Using KubePlus Platform toolkit,

* DevOps Engineer constructs a custom PaaS comprised of required Kubernetes Operators.

* Application Developer declares and creates application platforms as code leveraging custom resources
  introduced by the installed Operators.

We bring consistency of usage across multiple Operators with our Operator development guidelines_.
Teams can Build their Own PaaSes on Kubernetes selecting required Operators packaged as Helm charts.

.. _guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


.. image:: ./docs/KubePlus-Flow.jpg
   :scale: 75%
   :align: center


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

*--> No new CLI to learn*

KubePlus does not introduce any new CLI. Users can work with the same Kubernetes native interfaces like kubectl and YAML to leverage KubePlus functionality.


*--> Eliminate out-of-band platform automation*

Kubernetes Operators embed platform element life-cycle actions directly in Kubernetes. An example of a Kubernetes Operator can be Postgres Operator that 
embeds life-cycle actions such as create a database, add user to the database, change password of a user etc.
Such Operators leverage Kubernetes's strength of control loop (current state -> desired state) eliminating additional out-of-band automation.


*--> Uniformity between native and custom Kubernetes resources*

Our Operator development guidelines are designed to ensure custom resources become 
first-class entities of Kubernetes. 


*--> Discovery of custom resources*

KubePlus installs an additional component, KubePlus Discovery API Server, on your Kubernetes cluster to improve usability of custom Operators. This improves Platform-as-Code experience for application developers.


*--> Common language between Devs and Ops*

KubePlus leverages kubectl for management of Operators by Ops and their consumption by Devs. This makes Kubernetes YAMLs as the common language between Devs and Ops. 



KubePlus Architecture
======================

.. image:: ./docs/Architecture.jpg
   :scale: 75%
   :align: center

KubePlus consists of two components - Operator CRD and Discovery API Server.


1) Operator CRD:
----------------

For managing Operator life-cycle, KubePlus defines a Custom Resource of its own - named 'Operator'.
DevOps engineers create a Custom Resource instance of the 'Operator' Kind and include in its Spec the URL of the Operator Helm chart that they want to deploy. If you want to modify values of the Helm chart, specify them within the Operator Spec definition. For instance see example of `Moodle Operator definition`__.

.. _moodleoperator: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle/moodle-operator.yaml

__ moodleoperator_

If you want to deploy new version of an Operator, first delete the old version and then deploy the new version. 


2) Discovery API Server:
-------------------------

For helping application developers to create their application platforms as Code, KubePlus includes a Kubernetes Aggregated API server that exposes endpoints for application developers to find out more information about Custom Resources. Currently there are two endpoints - 'explain' and 'composition'. Examples of possible future endpoints are: 'provenance', 'functions', and 'configurables'. We look forward to inputs from the community on what additional information on Custom Resources you would like to get from such endpoints.

The 'explain' endpoint is used to get the documentation about Custom Kinds that are introduced by different Operators. This documentation is essentially OpenAPI Spec for that Custom Resource. It is supposed to be generated by Operator developers and packaged along with the Operator Helm Chart. In KubePlus we provide a tool based on kube-openapi library to generate OpenAPI Spec for your custom resources. 

The 'composition' endpoint is used by application developers for obtaining composition tree of native Kubernetes resources that are created as part of handling Custom Resources. We expect some of these endpoints to become part of upstream Kubernetes main API server eventually. For instance, `Kubernetes 1.13`__ includes 'explain' functionality for custom resources. We will deprecate such endpoints from KubePlus when the functionality becomes generally available for everyone.


.. _upstreamexplain: https://github.com/kubernetes/kubernetes/pull/67205

__ upstreamexplain_


Demo
====

Concept demo: https://youtu.be/Fbr1LNqvGRE

Working demo: https://drive.google.com/file/d/1jDptIWM8fiAorlZdW-pwOMttxAQAZHIR/view


Try it
=======

We provide three sample Operators that you can try - Postgres, Moodle, MySQL (derived from `Oracle MySQL Operator`__).

.. _oraclemysql: https://github.com/cloud-ark/mysql-operator

__ oraclemysql_

Postgres
---------

Follow steps in `examples/postgres/steps.txt`__.

.. _postgressteps: https://github.com/cloud-ark/kubeplus/blob/master/examples/postgres/steps.txt

__ postgressteps_


Moodle
-------

Follow steps in `examples/moodle/steps.txt`__.

.. _moodlesteps: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle/steps.txt

__ moodlesteps_


MySQL
-----

Follow steps in `examples/mysql/steps.txt`__.

.. _mysqlsteps: https://github.com/cloud-ark/kubeplus/blob/master/examples/mysql/steps.txt

__ mysqlsteps_


Multiple Operators
-------------------

Follow steps in `examples/multiple-operators/steps.txt`__.

.. _multipleoperatorssteps: https://github.com/cloud-ark/kubeplus/blob/master/examples/mysql/steps.txt

__ multipleoperatorssteps_




Quick try
-----------

Here is summary of deploying Postgres Operator.

KubePlus leverages Helm's Tiller component for deploying Operator charts.
So first you want to install Tiller.

**1) Install Helm/Tiller (by cluster administrator)**

  ``$ helm init``

Check Tiller Pod is ready

   ``$ kubectl get pods -n kube-system``

**2) Install KubePlus (by cluster administrator)**

  ``$ kubectl apply -f deploy/``

Check KubePlus is ready

  ``$ kubectl get pods``

KubePlus consists of 4 containers - operator-manager, operator-deployer, kube-discovery-apiserver, etcd.
KubePlus also deploys Tiller. 

Wait till all 4 KubePlus containers and Tiller Pod is in 'Running' state.


**3) Create custom PaaS (by cluster administrator)**


a) Once KubePlus is READY, Kubernetes cluster administrators define Kubernetes Operators to be installed in yaml files (e.g.: Postgres_, MySQL_, Moodle_) 
and use following kubectl commands:

.. _Postgres: https://github.com/cloud-ark/kubeplus/blob/master/examples/postgres/postgres-operator.yaml

.. _MySQL: https://github.com/cloud-ark/kubeplus/blob/master/examples/mysql/mysql-operator-chart-0.2.1.yaml

.. _Moodle: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle/moodle-operator.yaml


b) Deploy/install Operators:

  ``$ kubectl apply -f <operator yaml file>``


c) Find out all the installed Operators:

  ``$ kubectl get operators``


**4) Create Application Platform as Code (by application developer)**

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



Available Operators
====================

We are maintaining a `repository of Operators`__ that follow the guidelines. You can use Operators
from it or create your own Operator and use it with KubePlus. We can also help with checking
your Operators against the guidelines. Just open an issue on the repository with link to your Operator
code and we will provide you feedback on it.

.. _repository: https://github.com/cloud-ark/operatorcharts/

__ repository_


If you are interested in building your own operators, you can follow steps here_.

.. _here: https://github.com/cloud-ark/kubeplus/issues/14

You can also use tools like kubebuilder_ or `Operator SDK`__ to build your Operator.

.. _kubebuilder: https://github.com/kubernetes-sigs/kubebuilder

.. _sdk: https://github.com/operator-framework/operator-sdk

__ sdk_


Issues/Suggestions
===================

Follow `contributing guidelines`__ to submit bug reports or feature requests.

.. _contributing: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md

__ contributing_