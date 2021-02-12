.. CloudARK documentation master file, created by
   sphinx-quickstart on Wed Aug 30 10:11:27 2017.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Welcome to KubePlus documentation
====================================

KubePlus enables creating Kubernetes-style APIs from Helm charts with the ability to define policies and monitoring for the resources defined in the charts. 
Helm charts can contain any valid Kubernetes resource (built-in or custom). For custom resources defined in a chart, KubePlus assumes that the corresponding Operator is already installed on the cluster.

KubePlus simplifies creating multi-tenant SaaS application stacks with the required governance and monitoring defined per tenant. Think WordPress-as-a-Service with different Wordpress stacks created for different tenants, where the Wordpress stack is defined as a Helm chart. The second use case of KubePlus is when Platform Engineering teams want to share an opinionated Kubernetes workflow with their users. Platform teams can create a Kubernetes-style API wrapping the Helm chart representing such a  workflow and define required policies and monitoring for it. When users consume this API, KubePlus enables the Platform team to transaparently configure and monitor the resources that are created as part of the underlying Helm releases.

Next section provides details about KubePlus components.

.. toctree::
   :maxdepth: 3
   :caption: Contents:

   kubeplus-components
   operators
   community
   comparison
   contact
   getting-started
   faq

..   roadmap
..   examples
..   troubleshooting

..   architecture


