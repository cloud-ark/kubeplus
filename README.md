## KubePlus - Kubernetes Operator to deliver Helm charts as-a-service

Today Platform Engineering teams are dealing with a wide variety of Helm charts coming from different sources - open-source repositories, software vendors, Cloud marketplaces, or enterprise internal stakeholders. The desire is to deliver these custom application stacks as-a-service. Platform teams want an easy way for their customers to instantiate these application instances while being able to manage and monitor them behind the scene.

KubePlus offers a turn-key solution to deliver any Helm chart as-a-service. KubePlus takes an application Helm chart and delivers it as-a-service by abstracting it under provider and consumer APIs. The consumer of the service can simply use Kubernetes native consumer API to provision the application. And behind the scene provider of the service can apply policies and monitor the instances of the application.

KubePlus consists of a CRD that enables creating new Kubernetes APIs (CRDs) to realize such services. The new CRDs enable creation of a Helm release per tenant with tenant level isolation, monitoring and consumption tracking.

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

KubePlus is designed to help software vendors accelerate their journey to SaaS or enterprise platform teams rapidly deliver managed services for any custom applications.

KubePlus is part of the CNCF landscape's [Application Definition section](https://landscape.cncf.io/card-mode?category=application-definition-image-build&grouping=category).

## Overview

KubePlus takes an application Helm chart and delivers it as-a-service by automating multi-tenancy management and day2 operations. KubePlus consists of a CRD that enables creating new Kubernetes APIs (CRDs) to realize such services. Provider of the service has privileged access to the cluster and is able to create these services from Helm charts. Consumer of the service has limited access to the cluster and is able to use newly created service API / CRD and create an instance of the application. Behind the scene the provider is able to upgrade, monitor and govern this service.

<p align="center">
<img src="./docs/kubeplus-provider-consumer.png" width="600" height="250" class="center">
</p>

- Create: Create service for any application packaged as Helm chart.
- Govern: Tenant level policies for isolation and resource utilization.
- Monitor: Tenant level consumption metrics for CPU, memory, storage, network.
- Troubleshoot: Tenant level Kubernetes resource relationship graphs. 


## Components

KubePlus consists of two components:


### 1. In cluster component - CRD for CRDs to design platform services wrapping Helm charts

<p align="center">
<img src="./docs/crd-for-crds-2.jpg" width="700" height="300" class="center">
</p>


KubePlus offers a Kubernetes Custom Resource Definition (CRD) named ```ResourceComposition``` to 
- Compose platform services wrapping Helm charts
- Define policies (e.g. Node selection, CPU/Memory limits, etc.) for managing resources of the service
- Get aggregated CPU/Memory/Storage/Network metrics for the service instances
(in Prometheus format)

Here is the high-level structure of ResourceComposition CRD: 

<p align="center">
<img src="./docs/crd-for-crds-1.png" width="650" height="250" class="center">
</p>

To understand this further let us see how a multi-tenant platform service can be created from WordPress Deployment Helm chart. The Helm chart creates Wordpress Pod and uses in-cluster MySQL managed by a third-party MySQL Operator.
KubePlus takes this Helm chart and other policy and monitoring inputs through ResourceComposition CRD as shown below to deliver a new CRD for Wordpress Service.

<p align="center">
<img src="./docs/wordpress-saas.png" width="650" height="250" class="center">
</p>

Here is a new platform service named WordpressService. 

<p align="center">
<img src="./docs/wordpress-service-crd.png" width="650" height="250" class="center">
</p>

A new CRD named WordpressService has been created here using ResourceComposition.
The spec properties of the WordpressService Custom Resource are the attributes exposed via the underlying Helm chart's values.yaml. Here is a YAML definition to create a WordpressService instance for a tenant.

<p align="center">
<img src="./docs/wordpress-service-tenant1.png" width="650" height="250" class="center">
</p>


### 2. Client side Kubectl plugins for monitoring and troubleshooting

KubePlus kubectl plugins enable users to discover, monitor and troubleshoot service instances. The primary plugin is: ```kubectl connections```. It provides resource relationship information for Kubernetes resource instances (custom or built-in) with other resources (custom or built-in) by tracking owner references, labels, annotations, and spec properties. KubePlus constructs these Kubernetes resource relationship graphs at runtime. This provides it the ability to offer fine grained visibility and control over service instances.

Here is the resource relationship graph for WordpressService instance discovered using the ```kubectl connections``` command. 
```kubectl connections WordpressService wp-service-tenant1```.

<p align="center">
<img src="./docs/wordpress-service-connections.png" class="center">
</p>

We have additional plugins such as ```kubectl metrics``` and ```kubectl applogs``` that use resource relationship graphs to aggregate metrics and logs for service instances.


Details about these components are available [here](https://cloud-ark.github.io/kubeplus/docs/html/html/kubeplus-components.html).


## Try it

- Install KubePlus kubectl plugins. They can be used with any Kubernetes resource (built-in resources like Pod, Deployment, or custom resources like MysqlCluster).

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

- Install Helm v3 and install KubePlus in-cluser component using following command. KubePlus can be installed in any Namespace.

```
   $ KUBEPLUS_NS=default (or any namespace in which you want to install KubePlus)
   $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-0.2.2.tgz?raw=true" -n $KUBEPLUS_NS
```

- Try following examples:
  - [Hello World service](./examples/multitenancy/hello-world/steps.txt)
  - [Wordpress service](./examples/multitenancy/wordpress-mysqlcluster-stack/steps.txt)
  - [Mysql service](./examples/multitenancy/stacks/steps.txt)
  - [MongoDB service](./examples/multitenancy/mongodb-as-a-service/steps.md)
  - [Multiple teams](./examples/multitenancy/team/steps.txt) with applications deployed later

- Debug:
  ```
  - kubectl logs kubeplus $KUBEPLUS_NS -c crd-hook
  - kubectl logs kubeplus $KUBEPLUS_NS -c helmer
  - kubectl logs kubeplus $KUBEPLUS_NS -c platform-operator
  - kubectl logs kubeplus $KUBEPLUS_NS -c webhook-cert-setup
  ```

- Cleanup:
  ```
  - wget https://github.com/cloud-ark/kubeplus/raw/master/deploy/delete-kubeplus-components.sh
  - ./delete-kubeplus-components.sh
  ```

## Operator Maturity Model

As enterprise teams build their custom platforms using community or in house developed Operators, they need a set of guidelines for Operator readiness in multi-Operator and multi-tenant environments. We have developed the [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) for this purpose. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators for curating community Operators towards building their custom platforms.


## Presentations/Talks

1. [Being a good citizen of the Multi-Operator world, Kubecon NA 2020](https://www.youtube.com/watch?v=NEGs0GMJbCw&t=2s)

2. [Operators and Helm: It takes two to Tango, Helm Summit 2019](https://youtu.be/F_Dgz1V5Q2g)

3. [KubePlus presentation at community meetings (CNCF sig-app-delivery, Kubernetes sig-apps, Helm)](https://github.com/cloud-ark/kubeplus/blob/master/KubePlus-presentation.pdf)


## Contact

Submit issues on this repository or reach out to our team on [Slack](https://join.slack.com/t/cloudark/shared_invite/zt-2yp5o32u-sOq4ub21TvO_kYgY9ZfFfw).


## Status

Actively under development

