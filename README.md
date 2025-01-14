# KubePlus - Kubernetes Operator for Multi-Instance Multi-tenancy

## Intro

KubePlus is a turn-key solution that transforms any containerized application into a multi-instance SaaS.

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

Multi-instance multi-tenancy (MIMT) is a software architecture pattern in which a separate instance of an application is provided per tenant. The typical adopters of this pattern are application hosting providers, platform engineering teams, and B2B software vendors that need to host and manage dedicated instances of a software application for different tenants and effectively deliver that application as a managed service. KubePlus is a turn-key solution to build such managed services on Kubernetes.
It comes with end to end automation to help you deploy and manage your application on Kubernetes following the MIMT pattern. This includes isolation and security between instances along with easy to use APIs for managing upgrades, customization and resource utilization.

KubePlus takes an application Helm chart and wraps it under a Kubernetes API (CRD). Whenever an application instance is created using this API, KubePlus ensures that every instance is created in a separate namespace and the required multi-tenancy policies are applied in order to ensure isolation between instances. The API supports CRUD operations on the instances of the CRD, RBAC, version upgrades, and additional customizations for each instance.

<p align="center">
<img src="./docs/kubeplus-with-properties.png" width="700" height="250" class="center">
</p>

## Key Features

### Isolation

KubePlus takes an application Helm chart and wraps it in a Kubernetes API (CRD). This API is used to provision application instances on a cluster. KubePlus isolates each application instance in a separate namespace. It adds a safety perimeter around such namespaces using Kubernetes network policies and non-shared persistent volumes ensuring that each application instance is appropriately isolated from other instances.

### Security

The KubePlus Operator does not need any admin-level permissions on a cluster for application providers. This allows application providers to offer their managed services on any K8s clusters including those owned by their customers. KubePlus comes with a small utility that allows you to create provider specific kubeconfig on a cluster in order to enable application deployments and management. Providers have an ability to create a consumer specific further limited kubeconfig to allow for self-service provisioning of application instances as well.


### Resource Utilization

KubePlus provides controls to set per-namespace resource quotas. It also monitors usage of CPU, memory, storage, and network traffic at the application instance level. The collected metrics are available in different formats and can be pulled into Prometheus for historical usage tracking. KubePlus also supports ability to define licenses for the CRDs. A license defines the number of application instances that can be created for that CRD, and an expiry date. KubePlus prevents creation of application instances if the license terms are not met.

### Upgrades

A running application instance can be updated by making changes to the spec properties of the CRD instance and applying it.
KubePlus will update that application instance (i.e. helm upgrade of the corresponding helm release).
A new version of an application can be deployed by updating the application Helm chart under the existing Kubernetes CRD or registering the new chart under a new Kubernetes CRD. If the existing Kubernetes CRD object is updated, KubePlus will update all the running application instances (helm releases) to the new version of the application Helm chart.

### Customization

The spec properties of the Kubernetes CRD wrapping the application Helm chart are the fields defined in the chart’s values.yaml file. Application deployments can be customized by specifying different values for these spec properties.

## Installation

### Manual Installation Steps

1. **Create or use an existing Kubernetes cluster:**  
For testing purposes you can create a [minikube](https://minikube.sigs.k8s.io/docs/) or [kind](https://kind.sigs.k8s.io/) cluster:

   ```sh
   minikube start
   ```

   or

   ```sh
   kind create cluster
   ```

2. **Set the Namespace for KubePlus deployment:**

   ```sh
   export KUBEPLUS_NS=default
   ```

3. **Unzip KubePlus plugins and set up the PATH:**

   ```sh
   wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   tar -zxvf kubeplus-kubectl-plugins.tar.gz
   export KUBEPLUS_HOME=`pwd`
   export PATH=$KUBEPLUS_HOME/plugins:$PATH
   kubectl kubeplus commands
   ```

4. **Create the provider kubeconfig using the `provider-kubeconfig.py` utility:**

   ```sh
   wget https://raw.githubusercontent.com/cloud-ark/kubeplus/master/requirements.txt
   wget https://raw.githubusercontent.com/cloud-ark/kubeplus/master/provider-kubeconfig.py
   python3 -m venv venv
   source venv/bin/activate
   pip3 install -r requirements.txt
   apiserver=`kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'`
   python3 provider-kubeconfig.py -s $apiserver create $KUBEPLUS_NS
   deactivate
   ```

5. **Install KubePlus Operator using the generated provider kubeconfig:**

   ```sh
   helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-4.0.2.tgz?raw=true" --kubeconfig=kubeplus-saas-provider.json -n $KUBEPLUS_NS
   until kubectl get pods -A | grep kubeplus | grep Running; do echo "Waiting for KubePlus to start.."; sleep 1; done
   ```

### Automated Installation (Script)

The following script automates the steps outlined above for Linux systems,To install KubePlus and its kubectl plugin:

```sh
wget https://raw.githubusercontent.com/cloud-ark/kubeplus/master/install.sh
chmod +x install.sh
./install.sh --kubeplus --kubeplus-plugin
```

> **Note:** The script is a simplified method for installing KubePlus and its kubectl plugin. It combines the manual steps into a single execution for ease of use. It is recommended for those who want a quick setup without manually configuring each step.

## Demo

https://github.com/cloud-ark/kubeplus/assets/732525/efb255ff-fc73-446b-a583-4b89dbf61638



<!--
<p align="center">
<img src="./docs/app-metrics.png" width="700" height="250" class="center">
</p>-->

## To get started with an hands-on example

Follow: [Getting Started](examples/getting-started.md)
## Use Cases

KubePlus can be used for:

- [Application Hosting](./examples/multitenancy/application-hosting/wordpress/steps.txt)
- [Platform Engineering](./examples/multitenancy/platform-engineering/steps.txt)
- [Managed Service](./examples/multitenancy/managed-service/appday2ops/steps.txt)

To try these examples:

```sh
git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
cd kubeplus
export KUBEPLUS_HOME=`pwd`
export PATH=$KUBEPLUS_HOME/plugins:$PATH
```

Go to appropriate examples directory and follow `steps.txt` there in.
Make sure you have installed latest version of kubectl and you have created a minikube/kind cluster.

## Architecture

KubePlus architecture details are available [here](https://cloud-ark.github.io/kubeplus/docs/html/html/index.html).
KubePlus is a referenced solution for [multi-customer tenancy in Kubernetes](https://kubernetes.io/docs/concepts/security/multi-tenancy/#multi-customer-tenancy).

### Migrate to version 4.0.0+

If you are using KubePlus chart version < 4.0.0, follow these steps to migrate to 4.0.0+ versions.
In versions < 4.0.0, the KubePlus's built-in CRDs like `ResourceComposition` were included in the chart's
`templates` folder. This led to them getting deleted on KubePlus's upgrade. In 4.0.0+ version, the CRDs
have been moved to the `crds` folder, which avoids this issue.

```sh
kubectl annotate customresourcedefinition resourcepolicies.workflows.kubeplus helm.sh/resource-policy=keep
kubectl annotate customresourcedefinition resourceevents.workflows.kubeplus helm.sh/resource-policy=keep
kubectl annotate customresourcedefinition resourcemonitors.workflows.kubeplus helm.sh/resource-policy=keep
kubectl annotate customresourcedefinition resourcecompositions.workflows.kubeplus helm.sh/resource-policy=keep

helm upgrade kubeplus <4.0.0+ version> -n $KUBEPLUS_NS --kubeconfig=kubeplus-saas-provider.json
```

## Contributing

Check the [contributing guidelines](./Contributing.md).

## Case studies

1. [Bitnami Charts](https://cloudark.medium.com/kubeplus-verified-to-deliver-managed-services-with-100-bitnami-helm-charts-57eae3b9f6a6)

2. [Managed Jenkins Service at UT Austin](https://cloudark.medium.com/building-a-managed-jenkins-service-for-ut-austin-a-case-study-with-kubeplus-bdc082032f73)

## CNCF Landscape

KubePlus is part of CNCF landscape's
[Application Definition section](https://landscape.cncf.io/guide#app-definition-and-development--application-definition-image-build).

## Operator Maturity Model

As enterprise teams build their custom Kubernetes platforms using community or in house developed Operators, they need a set of guidelines for Operator readiness in multi-Operator and multi-tenant environments.
We have developed the [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) for this purpose. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready
 to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators for curating community Operators towards building their custom platforms.

## Presentations

1. [KubePlus presentation at community meetings (CNCF sig-app-delivery, Kubernetes sig-apps, Helm)](https://github.com/cloud-ark/kubeplus/blob/master/KubePlus-presentation.pdf)

2. [DevOps.com Webinar: Deliver your Kubernetes Applications as-a-Service](https://webinars.devops.com/deliver-your-kubernetes-applications-as-a-service)

3. [Being a good citizen of the Multi-Operator world, Kubecon NA 2020](https://www.youtube.com/watch?v=NEGs0GMJbCw&t=2s)

4. [Operators and Helm: It takes two to Tango, Helm Summit 2019](https://youtu.be/F_Dgz1V5Q2g)

## Community Meetings

We meet every Tuesday at 11.30 a.m. US CST. We use Slack huddle in `#kubeplus` channel on CNCF workspace
The meeting agenda is [here](https://docs.google.com/document/d/18PDo2XtvspP__3EemADyHh94O1-yActrLMCOntOiv1Y/edit?usp=sharing).
Please join us in our meetings. Your participation is welcome.

## Contact

Subscribe to [KubePlus mailing list](https://groups.google.com/g/kubeplus).

Join #kubeplus channel on [CNCF Slack](https://cloud-native.slack.com/archives/C06U6MP24PN).
If you don't have an account on the CNCF workspace, get your invitation [here](https://communityinviter.com/apps/cloud-native/cncf). You can join the `#kubeplus` channel once your invitation is active.
