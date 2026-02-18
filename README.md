# KubePlus - Kubernetes Operator for Multi-Instance Multi-tenancy

KubePlus is a Kubernetes Operator that enables teams to deliver applications as managed, multi-tenant services on Kubernetes. It bridges the gap between deploying an application and operating it at scale for multiple customers or internal teams, automating the isolation, access control, and lifecycle management that true multi-tenancy requires.

Many teams find that simply deploying a Helm chart is not enough when they need to offer the same application to many tenants or customers. Each tenant typically needs a separate instance with its own namespace, controlled access, resource limits, and upgrade paths. Building and maintaining this automation — and doing so safely for non-admin users — can require significant engineering effort.

KubePlus addresses these challenges by converting a Helm chart into a Kubernetes custom API (CRD) and managing the full lifecycle of each instance. When a user creates an instance of the custom resource, KubePlus creates a dedicated namespace, applies appropriate policies and quotas, deploys the underlying Helm release, and tracks all resources owned by that instance. This model enables safe delegation and operational visibility, while keeping everything within Kubernetes’ native API machinery.


## Intro

KubePlus is a turn-key Kubernetes Operator that transforms any containerized application packaged as a Helm chart into a managed, multi-tenant service. It implements the multi-instance multi-tenancy (MIMT) pattern, providing isolated application instances per tenant along with governance, policy enforcement, and lifecycle automation.

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

In the context of Kubernetes, multi-instance multi-tenancy (MIMT) means providing each tenant with its own isolated application instance, typically in a dedicated namespace. Unlike shared multi-tenant models where many tenants share the same application instance, MIMT ensures isolation, controlled access, and predictable resource usage. KubePlus implements the MIMT pattern by automating namespace creation, policy enforcement, RBAC mappings, and lifecycle operations for each tenant instance. The typical adopters of this pattern are application hosting providers, platform engineering teams, and B2B software vendors that need to host and manage dedicated instances of a software application for different tenants and effectively deliver that application as a managed service. 
KubePlus provides end-to-end automation to deploy and operate applications following the MIMT pattern on Kubernetes, including instance isolation, resource governance, RBAC enforcement, customization, and upgrades.


<p align="center">
<img src="./docs/kubeplus-with-properties.png" width="700" height="250" class="center">
</p>

## Key Features

### Isolation

KubePlus takes an application Helm chart and wraps it in a Kubernetes API (CRD). This API is used to provision application instances on a cluster. KubePlus isolates each application instance in a separate namespace. It adds a safety perimeter around such namespaces using Kubernetes network policies and non-shared persistent volumes ensuring that each application instance is appropriately isolated from other instances.

### Security

Because KubePlus creates custom APIs and controls instance provisioning, it enables service providers to delegate service operations without granting full cluster admin rights. This makes it practical to run managed services even on customer-owned clusters or shared environments. KubePlus comes with a small utility that allows you to create provider specific kubeconfig on a cluster in order to enable application deployments and management. Providers have an ability to create a consumer specific further limited kubeconfig to allow for self-service provisioning of application instances as well.


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
   Go to https://github.com/cloud-ark/kubeplus/releases
   Click "Assets" -> right click kubeplus-kubectl-plugins-v*.tar.gz and copy the link address
   wget "plugin link from above step"
   mkdir plugins
   mv kubeplus-kubectl-plugins-v*.tar.gz plugins/.
   cd plugins
   tar -zxvf kubeplus-kubectl-plugins-v*.tar.gz
   export PATH=`pwd`:$PATH
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
   python3 provider-kubeconfig.py -s $apiserver -x <cluster_name> create $KUBEPLUS_NS
   deactivate
   ```

5. **Install KubePlus Operator using the generated provider kubeconfig:**

   ```sh
   helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-4.2.0.tgz?raw=true" --kubeconfig=kubeplus-saas-provider.json -n $KUBEPLUS_NS
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

KubePlus supports a variety of production scenarios. Whether you are a platform team needing to deliver internal tools as self-service services, a hosting provider offering multiple instances of open source applications, or a software vendor building a SaaS offering on Kubernetes, KubePlus’ automation and isolation model simplifies operations.

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
