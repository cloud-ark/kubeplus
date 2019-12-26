============================================
Kubernetes API Add-on for Platform-as-Code 
============================================

Kubernetes Custom Resources and Custom Controllers, popularly known as `Operators`_, extend Kubernetes to run third-party softwares directly on Kubernetes. KubePlus API Add-on simplifies creation of platform workflows consisting of Custom and built-in resources. The main benefit of using KubePlus to application/microservice developers are:

- easily discover static and runtime information about Custom Resources available in their cluster
- easily define bindings between Custom and/or built-in Resources
- define dependency between Custom and/or built-in Resources in order to prevent out-of-order creation of resources in a workflow.

KubePlus API Add-on provides discovery endpoints, binding functions, and an orchestration mechanism to enable application developers to define platform workflows as code using Kubernetes Custom Resources.

You can think of KubePlus API Add-on as a tool that enables AWS CloudFormation/Terraform like experience when working with Kubernetes Custom Resources.

.. _Operators: https://coreos.com/operators/

.. _as Code: https://cloudark.io/platform-as-code


Discovery Endpoints
--------------------

Variety of static and runtime information is associated with Kubernetes Custom Resources.
This includes - Spec properties, usage information, implementation-level assumptions made by Operator developer,
composition tree of Kubernetes built-in resources created as part of handling Custom Resources, etc. 
KubePlus API Add-on defines following custom endpoints for static and runtime information discovery:

.. code-block:: bash

   kubectl get --raw "/apis/platform-as-code/v1/man"

.. code-block :: bash

   kubectl man <Custom Resource>

The man endpoint is used for obtaining static usage information about a Custom Resource. 

.. image:: ./docs/MysqlCluster-man-output.png
   :scale: 25%
   :align: center


.. code-block:: bash

   kubectl get --raw "/apis/platform-as-code/v1/composition"

.. code-block:: bash

   kubectl composition <Custom Resource> <Custom Resource Instance> [<Namespace]

The composition endpoint is used for obtaining runtime composition tree of Kubernetes built-in resources that are created as part of handling a Custom Resource instance.

.. image:: ./docs/MysqlCluster-composition-output.png
   :scale: 25%
   :align: center



Runtime Binding Functions
--------------------------

KubePlus API Add-on defines following functions that can be used to glue different Custom Resources together. 

.. code-block:: bash

   1. Fn::ImportValue(<Parameter>)

This function resolves the parameter value using runtime information in a cluster and imports that value into the Spec where the function is defined.

.. code-block:: bash

   1. Fn::AddLabel(label, <Resource>)

This function adds the specified label to the specified resource by resolving the resource name using runtime
information in a cluster.

Here is how the ``Fn::ImportValue()`` function can be used in a Custom Resource YAML definition.

.. image:: ./docs/mysql-cluster1.png
   :scale: 10%
   :align: left

.. image:: ./docs/moodle1.png
   :scale: 10%
   :align: right

In the above example the name of the ``Service`` object which is child of ``cluster1`` Custom Resource instance 
and whose name contains the string ``master`` is discovered at runtime and that value is injected as the value of
``mySQLServiceName`` attribute in the ``moodle1`` Custom Resource Spec.

Formal grammar of ``ImportValue`` and ``AddLabel`` functions is available in the `functions doc`_.

.. _functions doc: https://github.com/cloud-ark/kubeplus/blob/master/docs/kubeplus-functions.txt


Check our `slide deck`_ in the Kubernetes Community Meeting for more details of the above example.


PlatformStack Operator
-----------------------
Creating workflows requires treating the set of resources that representing the workflow as a unit.
For this purpose KubePlus provides a Custom Resource of its own - ``PlatformStack``. This Custom Resource enables application developers to define all the resources in a workflow as a unit along with the inter-dependencies between them. The dependency information is used to prevent out-of-order creation of resources. PlatformStack Operator does not actually deploy any resources defined in a workflow stack. Resource creation is done by application developers as usual using 'kubectl'.

.. image:: ./docs/platform-stack1.png
   :scale: 10%
   :align: center


KubePlus Components 
--------------------

Discovery endpoints, runtime binding functions and the PlatformStack Custom Resource are implemented using following components - an Aggregated API Server, a Mutating webhook, and an  Operator.

.. image:: ./docs/KubePlus-components1.jpg 
   :scale: 25% 
   :align: center

Additionally, KubePlus API Add-on defines following Platform-as-Code annotations. 

.. code-block:: bash

   platform-as-code/composition 

The 'composition' annotation is used to define Kubernetes's built-in resources that are created as part of instantiating a Custom Resource instance.

.. code-block:: bash

   platform-as-code/usage 

The 'usage' annotation is used to define usage information for a Custom Resource.
The value for 'usage' annotation is the name of the ConfigMap that stores the usage information.
These annotations need to be defined on the Custom Resource Definition (CRD) YAMLs of Operators
in order to make Custom Resources discoverable and usable by application developers.

As an example, annotations on MysqlCluster Custom Resource Definition (CRD) are shown below:

.. code-block:: yaml

  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: mysqlclusters.mysql.presslabs.org
    annotations:
      helm.sh/hook: crd-install
      platform-as-code/composition: StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget
      platform-as-code/usage: mysqlcluster-usage.usage
  spec:
    group: mysql.presslabs.org
    names:
      kind: MysqlCluster
      plural: mysqlclusters
      shortNames:
      - mysql
    scope: Namespaced


Getting started
----------------

Read our `blog post`_ to understand how Kubernetes Custom Resources affect the notion of 'as-Code' systems.

.. _blog post: https://medium.com/@cloudark/kubernetes-and-the-future-of-as-code-systems-b1b2de312742


Install KubePlus:

.. code-block:: bash

   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
   $ ./deploy-kubeplus.sh

Install KubePlus kubectl plugins:

We provide kubectl plugins for 'man' and 'composition' endpoints to simplify querying of the man page
information and composition tree information about Custom Resources. In order to use the plugins you
will need to add KubePlus folder to your PATH variable.

.. code-block:: bash

   $ export PATH=$PATH:`pwd`


Platform-as-Code examples:

1. `Manual discovery and binding`_

.. _Manual discovery and binding: https://github.com/cloud-ark/kubeplus/blob/master/examples/moodle-with-presslabs/steps.txt


2. `Automatic discovery and binding`_

.. _Automatic discovery and binding: https://github.com/cloud-ark/kubeplus/blob/master/examples/platform-crd/steps.txt


Operator Maturity Model
------------------------

In order to build Platform workflows as code using Operators and Custom Resources, it is important for Cluster
administrators to evaluate different Operators against a standard set of requirements. We have developed
`Operator Maturity Model`_ towards this focusing on Operator usage in increasingly complex scenarios.

.. _Operator Maturity Model: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


KubePlus API Add-on Stakeholders
---------------------------------

KubePlus API Add-on is useful to Operator developers, DevOps Engineers, and Application/Microservice developers alike.

.. image:: ./docs/Platform-as-Code-workflow.jpg
   :scale: 25%
   :align: center

.. _discoverability and interoperability guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md


*1. Operator Developer*

For Operator developers, we have developed `Operator Maturity Model`_ with specific focus on Operator interoperability in multi-Operator environments. Use these guidelines when developing your Operator to ensure that it works smoothly with other Operators in a cluster.


*2. DevOps Engineer/Cluster Administrator*

DevOps Engineers/Cluster Administrators use standard tools such as 'kubectl' or 'helm' to deploy required Operators in a Kubernetes cluster. Additionally, they deploy KubePlus API Add-on in their cluster to equip application developers to discover and use various Custom Resources efficiently. We are maintaining a `repository of Operator helm charts`_
where every Operator Helm chart is annotated with Platform-as-Code annotations. 
Use it for building your custom platform layer using Operators.

.. _repository of Operator helm charts: https://github.com/cloud-ark/operatorcharts/


*3. Application/Microservices Developer*

Application/Microservices Developers use KubePlus API Add-on discovery endpoints, runtime binding functions, and PlatformStack Operator to create their platform workflows as-code.


KubePlus in Action
-------------------

1. Kubernetes Community Meeting notes_

.. _notes: https://discuss.kubernetes.io/t/kubernetes-weekly-community-meeting-notes/35/60

2. Kubernetes Community Meeting `slide deck`_

.. _slide deck: https://drive.google.com/open?id=1fzRLBpCLYBZoMPQhKMQDM4KE5xUh6-xU

3. Kubernetes Community Meeting demo_

.. _demo: https://www.youtube.com/watch?v=taOrKGkZpEc&feature=youtu.be


Comparison
-----------

Check comparison of KubePlus with other `community tools`_.

.. _community tools: https://github.com/cloud-ark/kubeplus/blob/master/Comparison.md



Operator FAQ
-------------

New to Operators? Checkout `Operator FAQ`_.

.. _Operator FAQ: https://github.com/cloud-ark/kubeplus/blob/master/Operator-FAQ.md



Bug reports
------------

Follow `contributing guidelines`_ to submit bug reports.

.. _contributing guidelines: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


Status
-------
Actively under development.