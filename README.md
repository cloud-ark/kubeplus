## KubePlus - Kubernetes Operator to deliver Helm charts as-a-service

As enterprise adoption of Kubernetes is growing, we see multiple teams collaborate on a Kubernetes cluster to realize the broader organizational goals. Typically, there is one team that is offering a service that the other team is looking to consume. It can be an ISV offering a service for their data & analytics software that their customer is looking to consume or it can be a platform team offering a service for an internal application that the product team is planning to use (e.g secret management, data processing etc.). Such teams can be thought of as providers and consumers in the context of delivering and consuming software on Kubernetes. The softwares that providers want to enable their consumers to use are typically available in the form of a Helm chart.

KubePlus is a turn-key solution that enables provider teams to deliver any Helm chart as-a-service. KubePlus takes an application Helm chart and delivers it as a service by abstracting it under provider and consumer APIs.

KubePlus brings following advantages to provider teams:
- API-based access to Helm charts on a cluster for consumers.
- Seamless support for Namespace-based multi-tenancy where each application instance (Helm release) can be deployed in a separate namespace.
- Monitoring and governance of application instances. 
- Tracking consumption metrics (cpu, memory, storage and network) at Helm release / application level. (Providers can use these metrics to define consumption-based chargeback model.)


<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>


## Overview

At a high-level a provider is looking for:
- Ability to create consumer APIs through which consumers can provision the software packaged as Helm charts. The API should expose only minimum set of parameters that they want consumers to control when provisioning the software instance.
- Ability to troubleshoot and monitor a deployed instance of the software.
- Ability to track consumption of the software by different consumers.

KubePlus achieves these goals as follows.
The ```provider API``` in KubePlus is a built-in CRD (Custom Resource Definition) that enables registering application Helm charts by 
creating new Kubernetes APIs (CRDs). These new CRD are essentially the ```consumer API``` which the application consumers use to provision / instantiate the registered Helm chart in a self-service manner. As part of the Helm chart registration step, the provider team can define policies that KubePlus applies to every instantiation of the registered chart.

<p align="center">
<img src="./docs/provider-consumer.png" width="600" height="200" class="center">
</p>

KubePlus functions:
- Create: Create service for any application packaged as Helm chart.
- Govern: Tenant level policies for isolation and resource utilization.
- Monitor: Tenant level consumption metrics for cpu, memory, storage, network.
- Troubleshoot: Tenant level Kubernetes resource relationship graphs. 


## Example

To understand the working of KubePlus and provider/consumer APIs further, let us see how a multi-tenant platform service can be created from WordPress Helm chart. The Helm chart defines a Wordpress Pod and a MySQL managed by a third-party MySQL Operator.

### Provider actions:

- Provider team creates a new consumer API (CRD) named WordpressService using ```ResourceComposition``` CRD/provider API as shown below. 

<p align="center">
<img src="./docs/wordpress-service-crd.png" width="650" height="250" class="center">
</p>

The spec properties of the WordpressService Custom Resource are the attributes exposed via the WordPress Helm chart's values.yaml. 

- Provider team uses kubeplus kubectl plugins to troubleshoot or monitor WordpressService instances. Here is an example of using ```kubectl metrics``` plugin that shows cpu, memory, storage, network ingress/egress for a Wordpress application.
The metrics output is also available in promtheus format.

<p align="center">
<img src="./docs/wordpress-metrics-pretty.png" class="center">
</p>

We have additional plugins such as ```kubectl connections``` and ```kubectl applogs``` that are useful for tracking resource relationship graphs and obtaining logs for service instances.

### Consumer action

The consumer uses WordpressService CRD (Consumer API) to provision an instance of WordPress stack.
KubePlus includes a web portal through which the service instances can be created.
The portal runs on the cluster and is accessible through local proxy. We provide a script 
to connect to this portal (deploy/open-consumer-ui.sh).
The portal is specific to a service. It shows consumer API documentation, a form to provide
inputs for creating a service instance, monitoring data for the created instance, and 
its resource relationship graph.

<p align="center">
<img src="./docs/consumerui-apidoc1.png"  class="center">
</p>

<p align="center">
<img src="./docs/consumerui-form1.png" class="center">
</p>

<p align="center">
<img src="./docs/consumerui-instance-details1.png" class="center">
</p>


## Components

KubePlus consists of an Operator and kubectl plugins.

### 1. KubePlus Operator

The KubePlus Operator runs on the cluster and consists of a custom controller, a mutating webhook and the helmer module that handles Helm related functions.

<p align="center">
<img src="./docs/crd-for-crds-2.jpg" width="700" height="300" class="center">
</p>

The custom controller handles KubePlus CRDs, the primary amongst which is ```ResourceComposition```. It is used to:
- Define new CRDs (consumer APIs) wrapping Helm charts
- Define policies (e.g. node selection, cpu/memory limits, etc.) for managing resources of the service
- Get aggregated cpu/memory/storage/network metrics for the service instances (in Prometheus format)

The mutating webook and helmer module support the custom controller in delivering the KubePlus experience.


### 2. KubePlus kubectl plugins

KubePlus kubectl plugins enable providers to discover, monitor and troubleshoot application instances. The primary plugin is: ```kubectl connections```. It tracks resource relationships through owner references, labels, annotations, and spec properties. These relationships enable providers to gain fine grained visibility into running application instances  through resource relationship graphs. Additional plugins offer the ability to get aggregated consumption metrics (for cpu/memory/storage/network) or logs at the application instance level. There are plugins to extract provider or consumer specific kubeconfigs as well.

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

- Install Helm v3 and install KubePlus Operator using the following command. KubePlus Operator can be installed in any Namespace.

```
   $ KUBEPLUS_NS=default (or any namespace in which you want to install KubePlus)
   $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-0.2.3.tgz?raw=true" -n $KUBEPLUS_NS
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

## CNCF Landscape

KubePlus is part of the CNCF landscape's [Application Definition section](https://landscape.cncf.io/card-mode?category=application-definition-image-build&grouping=category).


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

