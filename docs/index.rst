.. CloudARK documentation master file, created by
   sphinx-quickstart on Wed Aug 30 10:11:27 2017.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Welcome to KubePlus documentation
====================================

KubePlus is a framework to create managed services from Helm charts. It consists of a Kubernetes Operator that enables Platform Engineering teams to create new Kubernetes CRDs wrapping Helm charts with policies and monitoring controls. Platform Engineering teams are able to offer a SaaS like experience for any application stack packaged as a Helm chart with this. It enables them to create a Helm release per tenant with tenant isolation, tenant level policy and tenant level consumption tracking.

Think WordPress-as-a-Service with different Wordpress stacks created for different tenants, where the Wordpress stack is defined as a Helm chart.

Here are primary use cases of KubePlus:

-Enterprise platform engineering teams delivering software stack as a service to their internal clients.

-ISVs delivering managed service for their software on any managed Kubernetes service on public clouds.

-ISVs accelerate building multi-tenant SaaS for their software on Kubernetes.



Next section provides details about KubePlus components.

.. toctree::
   :maxdepth: 3
   :caption: Contents:

   kubeplus-components
   getting-started
   operators
   community
   comparison

..   roadmap
..   examples
..   troubleshooting

..   architecture


