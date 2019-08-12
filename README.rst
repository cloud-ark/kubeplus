=================================================
K8s Custom Resource Discovery and Binding Add-on
=================================================

KubePlus Custom Resource Discovery and Binding Add-on enables discovery and binding of Kubernetes Custom Resources to build Platforms as-Code. You can think of it as a layer that enables AWS CloudFormation like experience when working with Kubernetes Custom Resources.

Kubernetes Custom Resource Definitions (CRDs), popularly known as `Operators`_, extend Kubernetes to run third-party softwares directly on Kubernetes. KubePlus Discovery and Binding Add-on helps application developers in creating platform stacks declaratively using Kubernetes Custom Resources.

.. _Operators: https://coreos.com/operators/

.. _platforms as code: https://cloudark.io/platform-as-code


What does it do?
=================

KubePlus Discovery and Binding Add-on enables two functions - discovery and automatic binding for Kubernetes Custom Resources.

*Discovery* - Variety of static and dynamic information is associated with Kubernetes resources.
Some examples are - Spec properties, usage, implementation-level assumption made by an Operator, 
composition tree of Kubernetes resources created as part of handling Custom Resources, permissions granted to the CRD/Operator Pod, whether Custom Resources are in use as part of a platform stack, history of declarative actions performed on resources, etc. KubePlus Discovery and Binding Add-on enables discovering this type of information about Custom resources directly through 'kubectl'.


*Binding* - Assembling multiple resources - built-in and Custom - to build platform stacks requires them to be bound/tied together in specific ways. In Kubernetes 'labels', 'label selectors' and name-based dns resolution satisfy the binding needs between built-in resources. However, when using Custom Resources from different Operators these built-in mechanisms are not sufficient. Correct binding may require setting Spec properties to specific values or orchestrating actions on multiple resources. KubePlus Discovery and Binding Add-on enables automating binding through a minimal language that can be used to glue together different Custom Resources through their YAML definitions.


Who is the target user of KubePlus?
====================================

KubePlus is useful to anyone who works with Kubernetes Custom Resources. These could be service developers, microservice developers, application developers, or devops engineers.


How does it work?
==================

KubePlus Discovery and Binding Add-on consists of - an aggregated API Server, a mutating webhook, and a platform operator. Additionally, KubePlus comes with a small language and a set of endpoints that help with composing Custom Resources together.

KubePlus Language
------------------

The main goal of KubePlus is to make it easy for Custom Resource users to define "stacks" of Custom Resources to achieve their end goals. Towards this we have defined a minimal language that can be used to glue different Custom Resources together. Currently the language supports just two functions:

.. code-block:: bash

   1. Fn::ImportValue(<Parameter>)

This function imports value of the specified parameter into the Spec where the function is defined.

.. code-block:: bash

   1. Fn::AddLabel(label, <Resource>)

This function adds the specified label to the specified resource.

Formal grammar of the language is available in the `language doc`_.

.. _language doc: https://github.com/cloud-ark/kubeplus/blob/master/docs/kubeplus-language.txt

.. .. image:: ./docs/KubePlus-diagram.png
..   :scale: 20%
..   :align: center

KubePlus Endpoints
-------------------

In order to perform discovery and binding, KubePlus defines following custom endpoints:

.. code-block:: bash

   kubectl get --raw "/apis/platform-as-code/v1/composition"

The composition endpoint is used for obtaining runtime composition tree of Kubernetes resource instances that are created as part of handling a Custom resource instance.

.. image:: ./docs/Moodle-composition.png
   :scale: 25%
   :align: center


.. code-block:: bash

   kubectl get --raw "/apis/platform-as-code/v1/man"

The man endpoint is used for obtaining usage information about a Custom Resource. It is a mechanism that an Operator developer can use to expose any assumptions or usage details that go beyond the Spec properties.

.. image:: ./docs/Moodle-man.png
   :scale: 25%
   :align: center


These endpoints can be used manually as well as programmatically. In fact, the ``composition`` endpoint is used
by KubePlus internally as part of handling the language constructs.


Platform-as-Code Annotations
-----------------------------

For correct working of above endpoints following annotations need to be defined on the Custom Resource Definition (CRD) YAMLs.

.. code-block:: bash

   platform-as-code/composition 

The 'composition' annotation is used to define Kubernetes's built-in resources that are created as part of instantiating a Custom Resource instance.

.. code-block:: bash

   platform-as-code/usage 

The 'usage' annotation is used to define usage information for a Custom Resource.
The value for 'usage' annotation is the name of the ConfigMap that stores the usage information.

As an example, annotations on Moodle Custom Resource Definition are shown below:

.. code-block:: yaml

   apiVersion: apiextensions.k8s.io/v1beta1
   kind: CustomResourceDefinition
   metadata:
     name: moodles.moodlecontroller.kubeplus
     annotations:
       platform-as-code/usage: moodle-operator-usage.usage
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


Read our `blog post`_ to understand the challenges and the architecture of KubePlus Discovery and Binding Add-on.

.. _blog post: https://medium.com/@cloudark/kubeplus-platform-toolkit-simplify-discovery-and-use-of-kubernetes-custom-resources-85f08851188f


Getting started
=================

Install KubePlus:

.. code-block:: bash

   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
   $ ./deploy-kubeplus.sh

Try:

1. `Manual discovery and binding`_

.. _Manual discovery and binding: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt


2. `Automatic discovery and binding`_

.. _Automatic discovery and binding: https://github.com/cloud-ark/kubeplus/blob/master/examples/platform-crd/steps.txt


Platform-as-Code Practice
===========================

.. _discoverability and interoperability guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


*1. Operator Developer*

Operator developers add above mentioned annotations on their CRD definitions. They also create the ConfigMaps with the required content. We have developed `discoverability and interoperability guidelines`_ to help with Operator development.

*2. DevOps Engineer*

DevOps Engineers/Cluster Administrators use standard tools such as 'kubectl' or 'helm' to deploy required Operators in a cluster. Additionally, they deploy KubePlus in their cluster to enable their Application developers discover and use various Custom Resources efficiently.


*3. Application Developer*

Application developers use Platform-as-Code endpoints to discover static and dynamic information about Custom Resources in their cluster. Using this information they can then build their platform stacks 
composing various Custom Resources together using the KubePlus language.


Demo
====

See KubePlus in action_.

.. _action: https://youtu.be/wj-orvFzUoM


Available Operators
====================

We are maintaining a `repository of Operator helm charts`_ in which Operator CRDs are annotated with Platform-as-Code annotations.

.. _repository of Operator helm charts: https://github.com/cloud-ark/operatorcharts/


Feedback
=========

We are actively looking for inputs from the community on following aspects:

1. Language constructs

   - What additional language constructs would you like to see in KubePlus language?
     File your suggestions as comments on `issue 319`_

.. _issue 319: https://github.com/cloud-ark/kubeplus/issues/319


2. Endpoints

   - What additional endpoints would you like to see in KubePlus API Server?
     File your suggestions as comments on `issue 320`_

.. _issue 320: https://github.com/cloud-ark/kubeplus/issues/320



Bug reports
============

Follow `contributing guidelines`_ to submit bug reports.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
=======

Actively under development.