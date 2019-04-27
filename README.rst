==========================
KubePlus Platform toolkit
==========================

KubePlus Platform toolkit simplifies discovery and use of Kubernetes Operators and Custom Resources in a cluster. Specifically, it supports discovering static as well as dynamic runtime information about Custom Resources. This information is necessary and useful when creating application platform stacks by assembling various Custom Resources together (platforms 'as code').

Kubernetes Custom Resource Definitions (CRDs), popularly known as `Operators`_, extend Kubernetes API to manage third-party software as native Kubernetes objects. Today, number of Operators are
being built for middlewares like databases, queues, ssl certificates, etc.
The Custom Resources introduced by Operators represent 'platform elements' -- their Spec definitions  encapsulate some higher-level workflow actions on the underlying infrastructure resource that they are managing (database, queue, ssl certificate, etc.). A novel approach for building platforms on Kubernetes is to construct a platform stack from multiple Custom Resources, essentially building the platform as Code. 

.. _Operators: https://coreos.com/operators/

The main challenge in this approach is interoperability between Custom Resources from different Operators. KubePlus Platform toolkit solves this challenge by standardizing on how application developers can easily discover and use various Custom Resources simply from 'kubectl' command-line. Specifically, it provides a way for application developers to obtain 'man page' like static information and 'pstree' like dynamic information for Custom Resources directly from 'kubectl'. Equipped with this information application developers can then create their platform stacks with correct YAML definitions of various Custom and native Kubernetes resources.


.. _guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


.. .. image:: ./docs/Kubeplus-flow-2.png
..   :scale: 25%
..   :align: center


Usage
======

KubePlus Platform toolkit is designed with 3 user personas in mind. 

*1. Operator Developer*

Operator developer uses our guidelines_ when developing their Operators. These guidelines enable creating Operators such that they are consistent to use alongside other Operators in a cluster.

*2. DevOps Engineer*

DevOps Engineer/Cluster Administrator uses Helm to deploy required Operators to create their custom PaaS. We provide_ Operators that adhere to our guidelines that you can use.

.. _provide: https://github.com/cloud-ark/operatorcharts/

*3. Application Developer*

Application developer uses kubectl to discover information about available Custom Resources
and then creates their application platforms as Code composing various Custom Resources together.


KubePlus Architecture
======================

KubePlus streamlines the process of discovering static and runtime information about Custom Resources
introduced by various Operators in a cluster. Static information consists of: a) how-to-use guides for Custom Resources supported by an Operator, b) any code level assumptions made by an Operator, c) OpenAPI Spec definitions for a Custom Resource. Runtime information consists of: a) identification of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance, 
b) historical information about declarative actions performed on Custom Resource instances.

KubePlus does not introduce any new input format, such as a new Spec, for Custom Resource discovery. Discovery is enabled via annotations, ConfigMaps, and custom sub-resources.

-----------------------------
Platform-as-Code Annotations
-----------------------------

KubePlus defines following annotations, which need to be set on Custom Resource Definitions (CRDs)

.. code-block:: bash

   platform-as-code/usage 

The 'usage' annotation is used to define usage information for a Custom Resource.

.. code-block:: bash

   platform-as-code/constants 

The 'constants' annotation is used to surface any code level assumptions made by an Operator.

.. code-blocK:: bash

   platform-as-code/openapispec 

The 'openapispec' annotation is used to define OpenAPI Spec for a Custom Resource.

The values for 'usage', 'constants', 'openapispec' annotations are names of ConfigMaps that store the corresponding data. Creating these ConfigMaps is the responsibility of Operator developer.
Don't forget to package these ConfigMaps along with your Operator Helm Chart. Here_ is an example of Moodle Helm Chart with these annotations and ConfigMaps.

.. _Here: https://github.com/cloud-ark/kubeplus-operators/tree/master/moodle/moodle-operator-chart/templates


.. code-block:: bash

   platform-as-code/composition 

The 'composition' annotation is used to define Kubernetes's native resources that are created by an Operator as part of instantiating instances of a Custom Resource. KubePlus uses this list, along with OwnerReferences, to build dynamic composition tree of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance.


Annotations on Moodle Custom Resource Definition are shown below:

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

Towards enabling application developers to discover information about Custom Resources directly from kubectl, KubePlus exposes following endpoints as custom sub-resources - 'man', 'explain' and 'composition'. 

These endpoints are implemented using Kubernetes's aggregated API Server.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"

The 'man' endpoint is used to find out 'man page' like information about Custom Resources.
It essentially exposes the information packaged in 'usage' and 'constants' annotations on a CRD.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle"  | python -m json.tool
   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"  | python -m json.tool

The 'explain' endpoint is used to discover Spec of Custom Resources. 
It exposes the information packaged in 'openapispec' annotation on a CRD.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Moodle&instance=moodle1&namespace=namespace1" | python -mjson.tool

The 'composition' endpoint is used by application developers for discovering the runtime composition tree of native Kubernetes resources that are created as part of provisioning a Custom Resource instance.
It uses listing of native resources available in 'composition' annotation and Custom Resource OwnerReferences to build this tree.

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
You can use Operators from this repository, or create your own Operator and use it with KubePlus. 
Make sure to add the platform-as-code annotations mentioned above to enable your Operator consumers to easily find static and runtime information about your Custom Resources right through kubectl.

.. _repository of Operators: https://github.com/cloud-ark/operatorcharts/



Issues/Suggestions
===================

Follow `contributing guidelines`_ to submit suggestions, bug reports or feature requests.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
=======

Actively under development.