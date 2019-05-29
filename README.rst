========================
KubePlus Cluster Add-on
========================

KubePlus Cluster Add-on simplifies discovery and use of Kubernetes Operators and Custom Resources.

Kubernetes Custom Resource Definitions (CRDs), popularly known as `Operators`_, extend Kubernetes to run third-party softwares (databases, queues, volume backup/restore, etc.) directly on Kubernetes. The Custom Resources introduced by an Operator essentially represent 'platform elements' as they encapsulate high-level workflow actions to be performed on the software that the Operator is managing. 
Entire platform stacks can be created by assembling multiple Custom Resources together (essentially, enabling `platforms as code`_ experience).

.. _Operators: https://coreos.com/operators/

.. _platforms as code: https://cloudark.io/platform-as-code


Read our `blog post`_ to know more about KubePlus Cluster Add-on.

.. _blog post: https://medium.com/@cloudark/kubeplus-platform-toolkit-simplify-discovery-and-use-of-kubernetes-custom-resources-85f08851188f


Custom Resource Interoperability
=================================

The main challenge in this platform-as-code approach is the interoperability between Custom Resources from different Operators. Specifically, following three binding related issues arise:

*a) Attribute-value based binding* - What should be the values of Spec attributes of different Custom Resources?

*b) Label-based binding* - Are there Kubernetes's native resources corresponding to Custom Resources to which labels need to be added in order for an Operator to function correctly? How to find such resources?

*c) Annotation-based binding* - Are there any specific annotations that need to be added on Custom or native Kubernetes resources for an Operator to function correctly?

KubePlus focuses on solving the interoperability challenge by standardizing on how Operator developers package information about their Custom Resources using Kubernetes-native mechanisms, and how Application developers can easily discover this information directly through kubectl.


Architecture
=============

KubePlus Cluster Add-on standardizes the process of defining static information about Custom Resources and discovering static and runtime dynamic information in Kubernetes-native manner. Static information consists of: a) how-to-use guides for Custom Resources, b) any code level assumptions made by an Operator towards handling a Custom Resource, c) OpenAPI Spec definitions for a Custom Resource. Runtime information consists of: a) identification of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance, b) history of declarative actions performed on Custom Resource instances.

KubePlus Cluster Add-on uses annotations, ConfigMaps, and custom endpoints to enable the discovery process.


.. .. image:: ./docs/KubePlus-diagram.png
..   :scale: 20%
..   :align: center


-----------------------------
Platform-as-Code Annotations
-----------------------------

KubePlus Cluster Add-on defines following annotations that need to be set on Custom Resource Definition (CRD) YAMLs.

.. code-block:: bash

   platform-as-code/usage 

The 'usage' annotation is used to define usage information for a Custom Resource.

.. code-block:: bash

   platform-as-code/constants 

The 'constants' annotation is used to define any code level assumptions made by an Operator.

.. code-blocK:: bash

   platform-as-code/openapispec 

The 'openapispec' annotation is used to define OpenAPI Spec for a Custom Resource.

The values for 'usage', 'constants', 'openapispec' annotations are names of ConfigMaps that store the corresponding data. 

.. code-block:: bash

   platform-as-code/composition 

The 'composition' annotation is used to define Kubernetes's native resources that are created as part of instantiating a Custom Resource instance. KubePlus Cluster Add-on uses the values in this annotation and OwnerReferences, to build dynamic composition tree of Kubernetes's native resources that are created as part of instantiating a Custom Resource instance.

As an example, annotations on Moodle Custom Resource Definition are shown below:

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

This Moodle CRD is part of the Moodle Operator whose Helm chart is available here_.

.. _here: https://github.com/cloud-ark/kubeplus-operators/tree/master/moodle/moodle-operator-chart/templates


----------------------------
Platform-as-Code Endpoints
----------------------------

For kubectl-based discovery, KubePlus Cluster Add-on exposes following endpoints - 'man', 'explain' and 'composition'. 

These endpoints are implemented using Kubernetes's aggregated API Server.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"

The 'man' endpoint is used to find out 'man page' like information about Custom Resources.
It essentially exposes the information packaged in 'usage' and 'constants' annotations.

.. image:: ./docs/Moodle-man.png
   :scale: 25%
   :align: center


.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle"  | python -m json.tool
   $ kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"  | python -m json.tool


The 'explain' endpoint is used to discover Spec of Custom Resources. 
It exposes the information packaged in 'openapispec' annotation.

.. image:: ./docs/Moodle-explain.png
   :scale: 25%
   :align: center


.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Moodle&instance=moodle1&namespace=namespace1" | python -mjson.tool


The 'composition' endpoint is used by Application developers for discovering the runtime composition tree of native Kubernetes resources that are created as part of provisioning a Custom Resource instance.
It uses listing of native resources available in 'composition' annotation and Custom Resource OwnerReferences to build this tree.

.. image:: ./docs/Moodle-composition.png
   :scale: 25%
   :align: center


Examples of possible future endpoints are: 'provenance', 'functions', and 'configurables'. We look forward to inputs from the community on what additional information on Custom Resources you would like to get from such endpoints.


Example of using KubePlus Cluster Add-on
=========================================

As an example of how KubePlus Cluster Add-on is useful, you can check the `Moodle Platform`_
built from three Operators — Moodle, MySQL, and Volume backup/restore. The various Custom Resources available through these Operators are — Moodle, MysqlCluster, Restic, Recovery. KubePlus helps application developers discover following aspects of these Custom Resources:

- Moodle Custom Resource YAML definition needs a specific value to bind to a MysqlCluster Custom Resource instance. Using the ‘man’ endpoint with Moodle and MysqlCluster Custom Resources as input helps here.

- In order to take backup of Moodle volume, the Deployment object for that Moodle Custom Resource instance needs to be given some label and that label needs to be used in the Restic Custom Resource label selector. The ‘man’ endpoint with Moodle and Restic as inputs help here. Also, the ‘composition’ endpoint with Moodle instance as input is needed to be used to find the name of the Deployment object.

- The Moodle volume backup also needs name of the Volume that needs to be backed up. The ‘man’ endpoint with Moodle Custom Resource input helps here as it surfaces the volume name which is an implementation detail of the Moodle Operator.

.. _Moodle Platform: https://github.com/cloud-ark/kubeplus/tree/master/examples/moodle-presslabs-stash


Usage
======

.. _discoverability and interoperability guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


*1. Operator Developer*

Operator developers add above mentioned annotations on their CRD definitions. They also create the ConfigMaps with the required content. We have developed `discoverability and interoperability guidelines`_ to help with Operator development.

*2. DevOps Engineer*

DevOps Engineers/Cluster Administrators use standard tools such as 'kubectl' or 'helm' to deploy required Operators in a cluster. Additionally, they deploy KubePlus Cluster Add-on in their cluster to enable their Application developers discover and use various Custom Resources efficiently.


*3. Application Developer*

Application developers use Platform-as-Code endpoints to discover static and runtime information about Custom Resources in their cluster. Using this information they can then build their platform stacks 
composing various Custom Resources together.



Demo
====

KubePlus Cluster Add-on in action_.

.. _action: https://youtu.be/wj-orvFzUoM


Try it
=======

Follow `these steps`_.

.. _these steps: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt


Available Operators
====================

We are maintaining a `repository of Operator helm charts`_ in which Operator CRDs are annotated with Platform-as-Code annotations.

.. _repository of Operator helm charts: https://github.com/cloud-ark/operatorcharts/


RoadMap
========

1. Automate the binding process between Custom Resources.
2. Working with Operator developers to define Platform-as-Code annotations on their Operators.
3. Integrating Kubeprovenance_ functionality into KubePlus Cluster Add-on.
4. Improving operator-analysis to check conformance of Operators with guidelines.
5. Tracking and visualizing entire platform stacks.

.. _Kubeprovenance: https://github.com/cloud-ark/kubeprovenance


Issues/Suggestions
===================

Follow `contributing guidelines`_ to submit suggestions, bug reports or feature requests.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
=======

Actively under development.