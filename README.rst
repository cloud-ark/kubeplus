==========================================
KubePlus API Discovery and Binding Add-on
==========================================

KubePlus API Discovery and Binding Add-on enables discovery and binding of Kubernetes in-built and Custom Resources to build Platforms as-Code.

Kubernetes Custom Resource Definitions (CRDs), popularly known as `Operators`_, extend Kubernetes to run third-party softwares directly on Kubernetes (databases, queues, volume backup/restore, etc.). Custom Resources introduced by an Operator essentially represent 'platform elements' as they encapsulate high-level workflow actions to be performed on the software that the Operator is managing.
Entire platform stacks can be created by assembling Kubernetes in-built and Custom Resources.

KubePlus API Discovery and Binding Add-on helps application developers in creating such platform stacks declaratively as Kubernetes YAML definitions.

.. _Operators: https://coreos.com/operators/

.. _platforms as code: https://cloudark.io/platform-as-code

Read our `blog post`_ to understand the challenges and the architecture of KubePlus API Discovery and Binding Add-on.

.. _blog post: https://medium.com/@cloudark/kubeplus-platform-toolkit-simplify-discovery-and-use-of-kubernetes-custom-resources-85f08851188f


What does it do?
=================

KubePlus API Discovery and Binding Add-on enables two functions over Kubernetes in-built and Custom Resource - discovery and automatic binding.

*Discovery* - This means discovering static and dynamic information about Kubernetes resources. 
Examples of static information include - Spec properties, usage examples, any implementation-level assumption made by a Operator, how-to-use guide for Custom Resources, etc.
In Kubernetes 'kubectl explain' command is available to discover Spec properties of in-built and Custom Resources.
But Spec properties typically don't include other kinds of static information mentioned above.
KubePlus API Discovery and Binding Add-on enables discovering this information through 'kubectl'.

Examples of dynamic information include - composition tree of Kubernetes objects created as part of handling in-built or Custom Resources, permissions granted to the CRD/Operator Pod, whether Custom Resources are in use as part of a platform stack, history of declarative actions performed on resources, etc. Similar to static information, KubePlus API Discovery and Binding Add-on enables discovering this dynamic information also through 'kubectl'.


*Binding* - Assembling multiple resources - in-built and Custom - to achieve different platform workflow actions requires them to be bound/tied together in specific ways. In Kubernetes 'labels', 'label selectors' and name-based dns resolution satisfy the binding needs between in-built resources. However, when using Custom Resources from different Operators these built-in mechanisms are not sufficient. Correct binding may require setting Spec properties to specific values or orchestrating actions on multiple resources.
KubePlus API Discovery and Binding Add-on enables automating binding through input/output variables defined on annotations and referenced in Spec properties.


Getting started
=================

Install KubePlus:
- git clone https://github.com/cloud-ark/kubeplus.git
- cd kubeplus
- kubectl apply -f deploy
- cd mutating-webhook
- make deploy


Working with in-built Resources
================================

You can use KubePlus Discovery and Binding Add-on even if you are not using CRDs/Operators. For in-built resources, you can use the 'composition' endpoint to discover the composition tree of in-built resources.

.. code-block:: bash

   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Deployment&instance=*&namespace=kube-system" | python -mjson.tool


Working with Custom Resources
==============================


1. `Manual discovery and binding resolution`_

.. _Manual discovery and binding resolution: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt


2. `Automatic binding resolution`_

.. _Automatic binding resolution: https://github.com/cloud-ark/kubeplus/blob/master/examples/automatic-binding-resolution/steps.txt


How does it work?
==================

KubePlus API Discovery and Binding Add-on uses annotations, ConfigMaps, and custom endpoints to enable discovery and binding. Following annotations need to be set on Custom Resource Definition (CRD) YAMLs.

.. .. image:: ./docs/KubePlus-diagram.png
..   :scale: 20%
..   :align: center


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


Platform-as-Code Practice
===========================

.. _discoverability and interoperability guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


*1. Operator Developer*

Operator developers add above mentioned annotations on their CRD definitions. They also create the ConfigMaps with the required content. We have developed `discoverability and interoperability guidelines`_ to help with Operator development.

*2. DevOps Engineer*

DevOps Engineers/Cluster Administrators use standard tools such as 'kubectl' or 'helm' to deploy required Operators in a cluster. Additionally, they deploy KubePlus API Discovery and Binding Add-on in their cluster to enable their Application developers discover and use various Custom Resources efficiently.


*3. Application Developer*

Application developers use Platform-as-Code endpoints to discover static and dynamic information about in-built and Custom Resources in their cluster. Using this information they can then build their platform stacks 
composing various Custom Resources together.



Demo
====

See KubePlus API Discovery and Binding Add-on in action_.

.. _action: https://youtu.be/wj-orvFzUoM



Available Operators
====================

We are maintaining a `repository of Operator helm charts`_ in which Operator CRDs are annotated with Platform-as-Code annotations.

.. _repository of Operator helm charts: https://github.com/cloud-ark/operatorcharts/


RoadMap
========

1. Working with Operator developers to define Platform-as-Code annotations on their Operators.
2. Integrating Kubeprovenance_ functionality into KubePlus Cluster Add-on.
3. Improving operator-analysis to check conformance of Operators with guidelines.
4. Tracking and visualizing entire platform stacks.

.. _Kubeprovenance: https://github.com/cloud-ark/kubeprovenance


Issues/Suggestions
===================

Follow `contributing guidelines`_ to submit suggestions, bug reports or feature requests.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
=======

Actively under development.