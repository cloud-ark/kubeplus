=========
KubePlus
=========

KubePlus Platform toolkit delivers `Platform as Code`__ experience on Kubernetes.
Platform-as-Code offers composable and repeatable way of creating application platforms
leveraging Kubernetes API extensions, also known as Operators. 

.. _pac: https://medium.com/@cloudark/evolution-of-paases-to-platform-as-code-in-kubernetes-world-74464b0013ca

__ pac_


Kubernetes Operators
=====================

`Kubernetes Operators`__ extend Kubernetes API to manage
third-party software as native Kubernetes objects. Today, number of Operators are
being built for middlewares like databases, queues, loggers, etc.
Current popular approach is to ‘self-assemble’ platform stacks using Kubernetes Operators of
choice. This requires significant efforts and there is 
lack of consistent user experience across multiple Operators.

.. _Operators: https://coreos.com/operators/

__ Operators_

KubePlus Platform toolkit streamlines the process of composing multiple Operators into a custom PaaS,
and enables creating application Platforms as Code. Using KubePlus Platform toolkit,

* DevOps Engineer constructs a custom PaaS comprised of required Kubernetes Operators.

* Application Developer declares and creates application platforms as code leveraging Custom Resources
  introduced by the installed Operators.

We bring consistency of usage across multiple Operators with our Operator development guidelines_.
Teams can Build their Own PaaSes on Kubernetes selecting required Operators packaged as Helm charts.

.. _guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


.. image:: ./docs/KubePlus-Flow.jpg
   :scale: 75%
   :align: center


KubePlus Usage
===============

KubePlus is designed with 3 user personas in mind. 

*1. Operator developer*

Operator developer/curator uses our guidelines_ when developing their Operator.

*2. Cluster Administrator*

Cluster Administrator uses Helm to deploy required Operators.

*3. Application developer*

Application developer uses kubectl to discover information about available Custom Resources
and then uses Kubernetes YAMLs to create their application platform stacks.

KubePlus does not introduce any new CLI. KubePlus users continue to use the
standard Kubernetes CLI (kubectl, helm) and YAML definition format to manage their platforms.

 
.. image:: ./docs/KubePlus-Platform-Kit.jpg
   :scale: 75%
   :align: center


Value of KubePlus
==================

*1. No new CLI*

KubePlus does not introduce any new CLI. Teams use kubectl and Kubernetes YAMLs to create platforms as Code.

*2. One platform experience*

Our Operator guidelines_ ensure one platform experience with uniform discovery, documentation and support across multiple Operators.

*3. Eliminate ad-hoc scripts*

Kubernetes Operators embed platform specific workflows as Kubernetes objects and eliminate out-of-band custom automation. 


KubePlus Architecture
======================

KubePlus primarily consists of an aggregated API Server that provides ability to discover static and runtime information about Custom Resources available in a Cluster. Towards helping application developers to create their application platforms as Code, we expose platform-as-code endpoints through this API server 
Three endpoints are supported currently -  'man', 'explain' and 'composition'. 

The 'man' endpoint provides capability to find 'man page' like information about Custom Resources.

The 'explain' endpoint is used to discover Spec of Custom Resources. 
In KubePlus we provide a tool based on kube-openapi library to generate OpenAPI Spec for your Custom Resources. 

The 'composition' endpoint is used by application developers for discovering the runtime composition tree of native Kubernetes resources that are created as part of provisioning Custom Resources. 

Examples of possible future endpoints are: 'provenance', 'functions', and 'configurables'. We look forward to inputs from the community on what additional information on Custom Resources you would like to get from such endpoints.

Demo
====

Concept demo: https://youtu.be/Fbr1LNqvGRE

Working demo: https://drive.google.com/file/d/1jDptIWM8fiAorlZdW-pwOMttxAQAZHIR/view


Try it
=======

Follow steps in `examples/moodle-with-presslabs/steps.txt`__.

.. _moodlesteps: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt

__ moodlesteps_



Available Operators
====================

We are maintaining a `repository of Operators`__ that follow the Operator development guidelines_. 
You can use Operators from this repository, or create your own Operator and use it with KubePlus. 
We can also help with checking your Operators against the guidelines. Just open an issue on the repository with link to your Operator code and we will provide you feedback on it.

.. _repository: https://github.com/cloud-ark/operatorcharts/

__ repository_



Issues/Suggestions
===================

Follow `contributing guidelines`__ to submit suggestions, bug reports or feature requests.

.. _contributing: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md

__ contributing_


Status
=======

Actively under development.