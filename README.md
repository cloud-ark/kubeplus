## KubePlus - Kubernetes SaaS Operator to deliver Helm charts as-a-service

KubePlus is a turn-key solution to transform any containerized application into a SaaS. It takes an application Helm chart and delivers it as-a-service by automating multi-tenancy management and day2 operations such as monitoring, troubleshooting and application upgrades. KubePlus consists of a CRD that enables creating new Kubernetes APIs (CRDs) to realize such services. The new CRDs enable creation of a Helm release per tenant with tenant level isolation, monitoring and consumption tracking.

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

KubePlus offers following benefits towards deploying a Kubernetes-native application (Helm chart) in SaaS form:
- Seamless support for Namespace-based multi-tenancy where each application instance (Helm release) is created in a separate namespace.
- Application-specific provider and consumer APIs for role based access to the clusters.
- Troubleshooting and governance of application instances.
- Tracking consumption metrics (cpu, memory, storage and network) at Helm release level in Prometheus. Application providers can use these metrics to define consumption-based chargeback models.

<p align="center">
<img src="./docs/jenkins-cpu-graph.png" class="center">
</p>


## Overview

The typical requirements in a service-based delivery model of Kubernetes applications are as follows:
- From cluster admin's perspective it is important to isolate different application instances from one another on the cluster.
- Application consumers need a self-service model to provision application instances.
- Application providers need to be able to troubleshoot application instances, monitor them, and track their resource consumption.

KubePlus achieves these goals as follows. KubePlus defines a ```provider API``` to create application-specific ```consumer APIs```.
The ```provider API``` is a KubePlus CRD (Custom Resource Definition) named ``ResourceComposition`` that enables registering an application Helm chart in the cluster by defining a new Kubernetes API (CRD) representing the chart. The new CRD is essentially the ```consumer API``` which the application consumers use to instantiate the registered Helm chart in a self-service manner. Through ``ResourceComposition``application providers can define application-level policies, which KubePlus applies when instantiating the registered chart as part of handling the consumer APIs.


<p align="center">
<img src="./docs/provider-consumer.png" width="600" height="200" class="center">
</p>

KubePlus offers following functions to application providers:
- Create: Create a Kubernetes-native API to represent an application packaged as a Helm chart.
- Govern: Define policies for isolation and resource utilization per application instance.
- Monitor: Track application-specific consumption metrics for cpu, memory, storage, network.
- Troubleshoot: Gain application-level insights through fine-grained Kubernetes resource relationship graphs.


## Demo

KubePlus comes with a control center with embedded Prometheus integration for providers to manage their SaaS across multiple Kubernetes clusters.
See the control center in [action](https://youtu.be/aIVnC4GKIV4).


## Under the hood of Provider/Consumer APIs

To understand the working of KubePlus, let us see how a Wordpress provider can offer a multi-tenant Wordpress service using KubePlus.


### Cluster admin actions

*1. Install KubePlus*

Cluster administrator installs KubePlus on their cluster.

```
$ KUBEPLUS_NS=default
$ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-2.0.8.tgz?raw=true" -n $KUBEPLUS_NS
```

*2. Retrieve Provider kubeconfig file*

KubePlus creates provider and consumer kubeconfig files with appropriately scoped
RBAC policies. Cluster admin needs to distribute them to application providers and consumers. KubePlus comes with kubectl plugins to retrieve these files.
The provider kubeconfig file has permissions to register application helm charts under consumer APIs in the cluster. The consumer kubeconfig file has permissions to perform CRUD operations on the registered consumer APIs.

```
$ kubectl retrieve kubeconfig provider $KUBEPLUS_NS > provider.conf
```

### Provider actions

*1. Create consumer API*

The provider team defines the consumer API named ```WordpressService``` using the ```ResourceComposition``` CRD (the provider API). The Wordpress Helm chart that underlies this service is created by the provider team. The spec properties of the ```WordpressService Custom Resource``` are the attributes defined in the Wordpress Helm chart's values.yaml.

As part of registering the consumer API, the provider team can define policies such as the cpu and memory that should be allocated to each Wordpress stack, or the specific worker node on which to deploy a Wordpress stack, etc. KubePlus will apply these policies to the Helm releases when instantiating the underlying Helm chart.

<p align="center">
<img src="./docs/wordpress-service-crd.png" width="650" height="250" class="center">
</p> 

[Here](https://raw.githubusercontent.com/cloud-ark/kubeplus/master/examples/multitenancy/wordpress-mysqlcluster-stack/wordpress-service-composition.yaml) is the ResourceComposition definition for the WordpressService.

*2. Grant permission to the consumer to create instances of WordpressService*

Before consumers can instantiate WordpressService resources, the provider needs to grant permission to the consumer for that resource. KubePlus comes with a kubectl plugin for this purpose:

```
kubectl grantpermission consumer wordpressservices provider.conf $KUBEPLUS_NS
```

*3. Provider team uses kubeplus kubectl plugins to troubleshoot and monitor WordpressService instances*.

Once a consumer has instantiated a WordpressService instance, the provider can
monitor and troubleshoot it using the various kubectl plugins that KubePlus provides.

With the ``kubectl connections`` plugin provider can check whether all Kubernetes resources have been created as expected. The graphical output makes it easy to check the connectivity between different resources.

```
kubectl connections WordpressService tenant1 default -k provider.conf -o png -i Namespace:default,ServiceAccount:default -n label,specproperty,envvariable,annotation 
```
<p align="center">
<img src="./examples/multitenancy/wordpress-mysqlcluster-stack/wp-tenant1.png" class="center">
</p>

Using ```kubectl metrics``` plugin, provider can check cpu, memory, storage, network ingress/egress for a WordpressService instance. The metrics output is available in pretty, json and Prometheus formats.

```
kubectl metrics WordpressService tenant1 default -o pretty -k provider.conf 
```

<p align="center">
<img src="./examples/multitenancy/wordpress-mysqlcluster-stack/wp-tenant1-metrics-pretty.png" class="center">
</p>

### Consumer action

The consumer uses WordpressService Custom Resource (the consumer API) to provision an instance of Wordpress stack. The instances can be created using ``kubectl`` or through a web portal.  Here is consumer portal for WordpressService showing the created ```tenant1``` instance.

<p align="center">
<img src="./examples/multitenancy/wordpress-mysqlcluster-stack/wp-tenant1-consumerui.png" class="center">
</p>



## Components

KubePlus consists of in-cluster components and components that run outside the cluster.

### 1. In-cluster components

The in-cluster components of KubePlus are the ``KubePlus Operator`` and the
``Consumer UI``.

The KubePlus Operator consists of a custom controller, a mutating webhook and the helmer module. Here is a brief summary of these components. Details about them are available [here](https://cloud-ark.github.io/kubeplus/docs/html/html/kubeplus-components.html).

<p align="center">
<img src="./docs/crd-for-crds-2.jpg" width="700" height="300" class="center">
</p>

The custom controller handles the ```ResourceComposition```. It is used to:
- Define new CRDs representing Helm charts (consumer APIs).
- Define policies (e.g. cpu/memory limits, node selection, etc.) for service instances.

The mutating webook and helmer modules support the custom controller in delivering the KubePlus experience.

The ``Consumer UI`` runs on the cluster and is accessible through proxy. Consumer UI is service specific and can be used to create service instances by consumers.


### 2. KubePlus components outside the cluster

The KubePlus components that run outside the cluster are: the KubePlus SaaS Manager control center and kubectl plugins.

The KubePlus SaaS Manager control center consists of ``Provider portal`` through which providers can manage their SaaS across different clusters. It comes with integrated Prometheus that enables tracking resource metrics for service instances.

KubePlus kubectl plugins enable providers to discover, monitor and troubleshoot application instances. The plugins track resource relationships through owner references, labels, annotations, and spec properties. These relationships enable providers to get aggregated consumption metrics (for cpu, memory, storage, network), and logs at the application instance level. The plugins are integrated within the Provider portal.


## Try

- Use Kubernetes version <= 1.20 and Helm version 3+. With minikube, you can create a cluster with a specific version like so:
```
    $ minikube start --kubernetes-version=v1.20.0
```

- Install KubePlus Operator.

```
   $ KUBEPLUS_NS=default (or any namespace in which you want to install KubePlus)
   $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-2.0.8.tgz?raw=true" -n $KUBEPLUS_NS
```

- Download and Install KubePlus SaaS Manager control center from [here](https://cloudark.io/download)

Unzip, untar kubeplus-saas-manager-control-center bundle and then follow the steps in the README.md therein. KubePlus SaaS Manager control center is currently supported for MacOS and Ubuntu.

```
   $ gunzip kubeplus-saas-manager-control-center.tar.gz
   $ tar -xvf kubeplus-saas-manager-control-center.tar
   $ cd kubeplus-saas-manager-control-center
   $ . ./install-kubeplus-control-center.sh
   $ ./start-control-center.sh
```

- Try following examples:
  - [Jenkins service](./examples/jenkins/non-operator/steps.txt)
  - [Hello World service](./examples/multitenancy/hello-world/steps.txt)
  - [Wordpress service](./examples/multitenancy/wordpress-mysqlcluster-stack/steps.txt)

- Debug:
  ```
  - kubectl logs kubeplus $KUBEPLUS_NS -c crd-hook
  - kubectl logs kubeplus $KUBEPLUS_NS -c helmer
  - kubectl logs kubeplus $KUBEPLUS_NS -c platform-operator
  - kubectl logs kubeplus $KUBEPLUS_NS -c webhook-cert-setup
  - kubectl logs kubeplus $KUBEPLUS_NS -c consumerui
  ```

- Cleanup:
  ```
  - helm delete kubeplus -n $KUBEPLUS_NS
  - wget https://github.com/cloud-ark/kubeplus/raw/master/deploy/delete-kubeplus-components.sh
  - ./delete-kubeplus-components.sh
  - cd kubeplus-saas-manager-control-center
  - ./stop-control-center.sh
  ```

## Kubectl plugins for discovery, monitoring and troubleshooting

KubePlus kubectl plugins enable discovery, monitoring and troubleshooting of Kubernetes applications. You can install them following these steps:

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```


## CNCF Landscape

KubePlus is part of CNCF landscape's [Application Definition section](https://landscape.cncf.io/card-mode?category=application-definition-image-build&grouping=category).


## Operator Maturity Model

As enterprise teams build their custom Kubernetes platforms using community or in house developed Operators, they need a set of guidelines for Operator readiness in multi-Operator and multi-tenant environments. We have developed the [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) for this purpose. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators for curating community Operators towards building their custom platforms.


## Presentations

1. [DevOps.com Webinar: Deliver your Kubernetes Applications as-a-Service](https://webinars.devops.com/deliver-your-kubernetes-applications-as-a-service)

2. [Being a good citizen of the Multi-Operator world, Kubecon NA 2020](https://www.youtube.com/watch?v=NEGs0GMJbCw&t=2s)

3. [Operators and Helm: It takes two to Tango, Helm Summit 2019](https://youtu.be/F_Dgz1V5Q2g)

4. [KubePlus presentation at community meetings (CNCF sig-app-delivery, Kubernetes sig-apps, Helm)](https://github.com/cloud-ark/kubeplus/blob/master/KubePlus-presentation.pdf)


## Contact

For support and new features [reach out to us](https://cloudark.io/kubeplus-saas-manager) or contact our team on [Slack](https://join.slack.com/t/cloudark/shared_invite/zt-2yp5o32u-sOq4ub21TvO_kYgY9ZfFfw).
