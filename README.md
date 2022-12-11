## KubePlus - Kubernetes Operator to create Multi-instance SaaS from Helm charts

KubePlus is a turn-key solution to transform any containerized application into a SaaS. It takes an application Helm chart and turns it into a SaaS by automating multi-tenancy management and day2 operations such as monitoring, troubleshooting and application upgrades. 

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

KubePlus offers following benefits towards deploying a Kubernetes-native application (Helm chart) in SaaS form:
- Seamless support for [Namespace-based multi-tenancy](https://kubernetes.io/docs/concepts/security/multi-tenancy/) where each application instance (Helm release) is created in a separate namespace.
- Tracking consumption metrics (cpu, memory, storage and network) at Helm release level in Prometheus. Application providers can use these metrics to define consumption-based chargeback models.
- Application-specific provider and consumer APIs for role-based access to application instances.
- Troubleshooting and governance of application instances.


## Overview

The typical requirements in creating Kubernetes-based SaaS are as follows:
- Isolate different application instances from one another on the cluster.
- Easily govern, monitor, troubleshoot, and upgrade application instances. 
- Enable self-service provisioning of application instances.

KubePlus achieves these goals as follows. KubePlus defines a Custom Resource (```provider API```) to create application-specific ```consumer APIs```.
The ```provider API``` (named ``ResourceComposition``) enables registering an application Helm chart in the cluster by defining a new Kubernetes API (CRD) representing the chart. The new CRD is the ```consumer API``` which the application consumers use to instantiate the registered Helm chart in a self-service manner. Through ``ResourceComposition``application providers can define application-level policies, such as cpu and memory to be provided for application Pods. KubePlus enforces these policies when instantiating the registered chart as part of handling the consumer APIs. The architecture details of KubePlus are available [here](https://cloud-ark.github.io/kubeplus/docs/html/html/index.html).

<p align="center">
<img src="./docs/provider-consumer.png" width="600" height="200" class="center">
</p>

We have built a [control center](https://cloudark.io/kubeplus-saas-manager) to manage application SaaS across multiple Kubernetes clusters. The control center contains embedded Prometheus to track application resource consumption. You can host the control center yourself. For support, you can purchase our [subscription](https://cloudark.io/contact). See the control center in action [here](https://youtu.be/ZVhTE6WSjVI).

## Try

- Create a minikube cluster:
```
    $ minikube start --kubernetes-version=v1.24.3
```

- Install KubePlus Operator and retrieve provider and consumer kubeconfig.
KubePlus generates kubeconfig files for providers and consumers.
Use the provider kubeconfig to register the cluster in the control center. 
```
   $ KUBEPLUS_NS=default
   $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-3.0.2.tgz?raw=true" -n $KUBEPLUS_NS
   $ kubectl get configmaps kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-provider\.json}" > provider.conf
   $ kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-consumer\.json}" > consumer.conf
```

- Start the control center. For the first time you will be asked to setup user credentials.
```
Usage: ./kubeplus-control-center.sh <start|stop> <http|https> <domain_name> [<inet_ip>]
```
```
   $ ./kubeplus-control-center.sh start http localhost localhost
```

<p align="center">
<img src="./docs/kubeplus-saas-manager-login-screen.png" style="width:75%; height:75%" class="center">
</p>


- Register the cluster by adding the provider kubeconfig 

<p align="center">
<img src="./docs/kubeplus-saas-manager-register-cluster.png" style="width:75%; height:75%" class="center">
</p>

- Register Wordpress Service

Use following helm chart:
```
https://github.com/cloud-ark/k8s-workshop/blob/master/wordpress-deployment-chart/wordpress-chart-0.0.3.tgz?raw=true
```

<p align="center">
<img src="./docs/kubeplus-saas-manager-register-service.png" style="width:75%; height:75%" class="center">
</p>

- Add the service to the cluster

<p align="center">
<img src="./docs/kubeplus-saas-manager-add-service-to-cluster.png" style="width:75%; height:75%" class="center">
</p>

- Create application instance

Either the provider kubeconfig or consumer kubeconfig can be used to instantiate application.
Here we show instantiating the application from the control center (which contains the provider kubeconfig).
You can distribute the consumer kubeconfig to your customer teams for self-service creation of application instances.

<p align="center">
<img src="./docs/kubeplus-saas-manager-create-instance1.png" style="width:75%; height:75%" class="center">
</p>


- Check application resource consumption in the Prometheus

<p align="center">
<img src="./docs/kubeplus-saas-manager-running-instance.png" style="width:75%; height:75%" class="center">
</p>


<p align="center">
<img src="./docs/kubeplus-saas-manager-prometheus.png" style="width:75%; height:75%" class="center">
</p>


- Check application topology

<p align="center">
<img src="./docs/kubeplus-saas-manager-topology.png" style="width:75%; height:75%" class="center">
</p>


- Troubleshoot application resources 

<p align="center">
<img src="./docs/kubeplus-saas-manager-kubectl-access.png" style="width:75%; height:75%" class="center">
</p>

## Troubleshoot KubePlus

```
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c kubeconfiggenerator
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c crd-hook
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c helmer
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c platform-operator
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c webhook-cert-setup
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c consumerui
```

- Cleanup:
  ```
  - helm delete kubeplus -n $KUBEPLUS_NS
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
