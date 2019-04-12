==========================
KubePlus Platform toolkit
==========================

KubePlus Platform toolkit enables discovery of Kubernetes Operators installed in a Cluster
and the Custom Resources they support.

Application developers use KubePlus in building their application platforms as Code.

Platform-as-Code offers composable and repeatable way of creating application platforms
leveraging Kubernetes API extensions, also known as Operators. 

.. _pac: https://medium.com/@cloudark/evolution-of-paases-to-platform-as-code-in-kubernetes-world-74464b0013ca

__ pac_


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

*1. Operator Developer/Curator*

Operator developer/curator uses our guidelines_ when developing their Operator.
This includes - creating Helm chart for the Operator, registering CRD as part of the Helm chart,
and adding Platform-as-Code annotations on their CRD definition.

*2. DevOps Engineer*

DevOps Engineers/Cluster Administrators use Helm to deploy required Operators.

*3. Application Developer*

Application developer uses kubectl to discover information about available Custom Resources
and then uses Kubernetes YAMLs to create their application platform stacks.


Design Philosophy
==================

When developing KubePlus we have following these two design principles:

*1. No new CLI*

KubePlus does not introduce any new CLI.
Teams continue to use standard CLIs (kubectl, helm) and YAML definition format to manage their platforms.

*1. No new input format*

KubePlus does not introduce any new input format (such as a new Spec).
Any and all the information is packaged using ConfigMaps.

Our Operator guidelines_ ensure one platform experience with uniform discovery, documentation and usage across multiple Operators.


KubePlus Architecture
======================

KubePlus streamlines the process around discovering static and runtime information about Custom Resources
introduced by installed Operators in a cluster. Static information consists of: a) how-to-use guides for a Custom Resource, b) any code level assumptions made by an Operator, c) OpenAPI Spec definitions for a Custom Resource, etc. Runtime information consists of Kubernetes's native resources created as part of instantiating a Custom Resource instance.

-----------------------------
Platform-as-Code Annotations
-----------------------------

KubePlus uses annotations as a mechanism to include static and dynamic information as part of Custom Resource Definition. These annotations on Moodle Custom Resource Definition are shown below:

.. code-block:: yaml

   apiVersion: apiextensions.k8s.io/v1beta1
   kind: CustomResourceDefinition
   metadata:
     name: moodles.moodlecontroller.kubeplus
     annotations:
       platform-as-code/usage: moodle-operator-usage.usage
       platform-as-code/constants: moodle-operator-implementation-details.implementation_choices
       platform-as-code/openapispec: moodle-openapispec.openapispec
       platform-as-code/composition: Deployment, Service, PersistentVolume, PersistentVolumeClaim, Secret, Ingress
   spec:
     group: moodlecontroller.kubeplus
     version: v1
     names:
       kind: Moodle
       plural: moodles
     scope: Namespaced


.. code-block:: bash
   platform-as-code/usage 

The 'usage' annotation is used to define usage information for the Custom Resource.

.. code-block:: bash
   platform-as-code/constants 

The 'constants' annotation is used to surface any code level assumptions that an Operator might have made.

.. code-blocK:: bash
   platform-as-code/openapispec 

The 'openapispec' annotation is used to define OpenAPI Spec for the Custom Resource.

.. code-block:: bash
   platform-as-code/composition 

The 'composition' annotation is used to define Kubernetes's native objects that are created by the Operator as part of instantiating instances of that Custom Resource.

The values for 'usage', 'constants', 'openapispec' annotations are names of ConfigMaps that store the corresponding data. Creating these ConfigMaps is the responsibility of Operator developer/curator.
Don't forget to package these ConfigMaps along with your Helm Chart. Here is example of Moodle_ Helm Chart
with these annotations and ConfigMaps.

.. _Moodle: https://github.com/cloud-ark/kubeplus-operators/tree/master/moodle/moodle-operator-chart/templates

The values in 'composition' annotation are used by KubePlus in building dynamic composition tree of Kubernetes's native resources that are created as part of instantiating a Custom Resource.


----------------------------
Platform-as-Code Endpoints
----------------------------

To make it easy for application developers to discover static and runtime information about Custom Resources in a cluster, KubePlus exposes following endpoints as custom subresources - 'man', 'explain' and 'composition'. 

These endpoints are implemented using Kubernetes's aggregated API Server. 

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"

The 'man' endpoint provides capability to find 'man page' like information about Custom Resources.
It essentially exposes the information packaged in 'usage' and 'constants' annotations.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle"  | python -m json.tool
   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"  | python -m json.tool

The 'explain' endpoint is used to discover Spec of Custom Resources. 
It exposes the information packaged in 'openapispec' annotation. 

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Moodle&instance=moodle1&namespace=namespace1" | python -mjson.tool

The 'composition' endpoint is used by application developers for discovering the runtime composition tree of native Kubernetes resources that are created as part of provisioning Custom Resources.
It uses listing of native resources available in 'composition' annotation, along with OwnerReferences, to build this tree.

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
Make sure to add the platform-as-code annotations mentioned above to enable your Operator consumers to easily find static and runtime information about your Custom Resources right through kubectl.

We can also help checking your Operators against the guidelines. Just open an issue on the repository with link to your Operator code and we will provide you feedback on it.

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