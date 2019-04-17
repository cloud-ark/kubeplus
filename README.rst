==========================
KubePlus Platform toolkit
==========================

KubePlus Platform toolkit simplifies discovery and use of Kubernetes Operators and their Custom Resources to create application Platforms as Code. It supports discovering static as well as dynamic runtime information about Custom Resources introduced by various Operators in your Cluster.

`Kubernetes Operators`_ extend Kubernetes API to manage
third-party software as native Kubernetes objects. Today, number of Operators are
being built for middlewares like databases, queues, loggers, etc.
Current popular approach is to ‘self-assemble’ platform stacks using Kubernetes Operators of
choice. This requires significant efforts to discover capabilities of each Operator and the Custom Resources that it supports; and there is lack of consistent user experience across multiple Operators.

.. _Kubernetes Operators: https://coreos.com/operators/


KubePlus Platform toolkit, a) provides Operator developers a set of Operator development guidelines_ to enable creating Operators such as they are consistent to use alongside other Operators, b) leverages standard installation tools like Helm to install Operators, and c) provides kubectl based mechanisms for application developers to discover static and runtime information about Custom Resources.
Application developers can create their application platform stack using various Custom Resources.

.. _guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


.. image:: ./docs/KubePlus-Flow.jpg
   :scale: 75%
   :align: center


Usage
======

KubePlus Platform toolkit is designed with 3 user personas in mind. 

*1. Operator Developer/Curator*

Operator developer/curator uses our guidelines_ when developing their Operator.

*2. DevOps Engineer*

DevOps Engineer/Cluster Administrator uses Helm to deploy required Operators to create their custom PaaS. We provide_ Operators that adhere to our guidelines_ that you can use.

.. _provide: https://github.com/cloud-ark/operatorcharts/

*3. Application Developer*

Application developer uses kubectl to discover information about available Custom Resources
and then creates their application platforms as Code composing various Custom Resources.


Design Philosophy
==================

When developing KubePlus Platform toolkit we have followed these design principles:

*1. No new input format*

KubePlus does not introduce any new input format, such as a new Spec, to enable Operator installation
or Custom Resource discovery. Operator installation is done in standard way using Helm.
Custom Resource information discovery is enabled via annotations, ConfigMaps, and custom sub-resources using Kubernetes aggregated API server. 


*2. No new CLI*

KubePlus does not introduce any new CLI.
Teams continue to use standard CLIs such as 'kubectl' and 'helm' to build and manage their platforms.


KubePlus Architecture
======================

KubePlus streamlines the process of discovering static and runtime information about Custom Resources
introduced by various Operators in a cluster. Static information consists of: a) how-to-use guides for Custom Resources supported by an Operator, b) any code level assumptions made by an Operator, c) OpenAPI Spec definitions for a Custom Resource. An example of runtime information is identification of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance.

-----------------------------
Platform-as-Code Annotations
-----------------------------

KubePlus uses annotations on Custom Resource Definitions to include the required information.
Following annotations are currently supported:

.. code-block:: bash

   platform-as-code/usage 

The 'usage' annotation is used to define usage information for a Custom Resource.

.. code-block:: bash

   platform-as-code/constants 

The 'constants' annotation is used to surface any code level assumptions made by an Operator.

.. code-blocK:: bash

   platform-as-code/openapispec 

The 'openapispec' annotation is used to define OpenAPI Spec for a Custom Resource.

.. code-block:: bash

   platform-as-code/composition 

The 'composition' annotation is used to list Kubernetes's native resoruces that are created by an Operator as part of instantiating instances of a Custom Resource.

The values for 'usage', 'constants', 'openapispec' annotations are names of ConfigMaps that store the corresponding data. Creating these ConfigMaps is the responsibility of Operator developer/curator.
Don't forget to package these ConfigMaps along with your Helm Chart. Here is example of Moodle_ Helm Chart with these annotations and ConfigMaps.

.. _Moodle: https://github.com/cloud-ark/kubeplus-operators/tree/master/moodle/moodle-operator-chart/templates

The values in 'composition' annotation are used by KubePlus in building dynamic composition tree of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance.


As an example, these annotations on Moodle Custom Resource Definition are shown below:

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


----------------------------
Platform-as-Code Endpoints
----------------------------

For application developers our goal is to be able to find out information about Custom Resources directly using kubectl -- similar to how Unix man pages work.

Towards that end KubePlus exposes following endpoints as custom sub-resources - 'man', 'explain' and 'composition'. 

These endpoints are implemented using Kubernetes's aggregated API Server.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"

The 'man' endpoint provides capability to find 'man page' like information about Custom Resources.
It essentially exposes the information packaged in 'usage' and 'constants' annotations on a CRD.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle"  | python -m json.tool
   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"  | python -m json.tool

The 'explain' endpoint is used to discover Spec of Custom Resources. 
It exposes the information packaged in 'openapispec' annotation on a CRD.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Moodle&instance=moodle1&namespace=namespace1" | python -mjson.tool

The 'composition' endpoint is used by application developers for discovering the runtime composition tree of native Kubernetes resources that are created as part of provisioning a Custom Resource instance.
It uses listing of native resources available in 'composition' annotation and OwnerReferences to build this tree.

Examples of possible future endpoints are: 'provenance', 'functions', and 'configurables'. We look forward to inputs from the community on what additional information on Custom Resources you would like to get from such endpoints.

Demo
====

See KubePlus in action_.

.. _action: https://youtu.be/wj-orvFzUoM


Try it
=======

Follow `these steps`_.

.. _these steps: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt


Available Operators
====================

We are maintaining a `repository of Operators`_ that follow the Operator development guidelines_. 
You can use Operators from this repository or create your own Operator and use it with KubePlus. 
Make sure to add the platform-as-code annotations mentioned above to enable your Operator consumers to easily find static and runtime information about your Custom Resources right through kubectl.

We can also help checking your Operators against the guidelines. Just open an issue on the repository with link to your Operator code and we will provide you feedback on it.

.. _repository of Operators: https://github.com/cloud-ark/operatorcharts/



Issues/Suggestions
===================

Follow `contributing guidelines`_ to submit suggestions, bug reports or feature requests.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
=======

Actively under development.