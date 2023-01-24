## KubePlus - Kubernetes Operator for Multi-Instance Multi-tenancy

KubePlus is a turn-key solution to transform any containerized application into a SaaS.

<p align="center">
<img src="./docs/application-stacks-1.png" width="700" height="150" class="center">
</p>

Multi-instance multi-tenancy (MIMT) is a software architecture pattern in which a dedicated instance of an application is provided per tenant. The typical adopters of this pattern are organizations that need to host and manage multiple instances of a software application for different tenants and effectively deliver that application as-a-service. KubePlus is a turn-key solution to build a multi-instance multi-tenancy (MIMT) pattern on Kubernetes and comes with end to end automation to help you deploy and manage your MIMT application on Kubernetes. This includes isolation and security between instances along with easy to use APIs for managing upgrades, customization and resource utilization. 

KubePlus takes an application Helm chart and wraps it under a Kubernetes API (CRD). Whenever an application instance is created using this API, KubePlus ensures that every instance is created in a separate namespace and the required multi-tenancy policies are applied in order to ensure isolation between instances. The API also supports RBAC, version upgrades and additional customizations for each instance. 


<p align="center">
<img src="./docs/kubeplus-with-properties.png" width="700" height="250" class="center">
</p>


### Isolation

KubePlus takes an application Helm chart and wraps it in a Kubernetes API. Application providers use this API to provision application instances on a cluster. KubePlus isolates each application instance in a separate Namespace. It adds a safety perimeter around such Namespaces using Kubernetes Network Policies and non-shared persistent volumes ensuring that each application instance is appropriately isolated from other instances. Additionally, it provides controls for application providers to deploy different tenant application instances on different worker nodes for node isolation. 

### Security

The KubePlus Operator does not need any admin-level permissions on a cluster for application providers. This allows application providers to offer their managed services on any K8s clusters including those owned by their customers. KubePlus comes with a small utility that allows you to create provider specific kubeconfig on a cluster in order to enable this RBAC. Application providers have an ability to create a consumer specific further limited kubeconfig to allow for self service provisioning of the instance as well. 

### Resource utilization
KubePlus provides controls to set per-namespace resource quotas. It also monitors usage of CPU, memory, storage, and network traffic at the application instance level. The collected metrics are available in different formats and can be pulled into Prometheus for historical usage tracking.

### Upgrades
A new version of an application can be deployed by updating the application Helm chart under the existing Kubernetes API or registering the new chart under a new Kubernetes API. If the existing Kubernetes API object is updated, KubePlus will deploy the new application instances using the new version of the application Helm chart.

### Customization
The spec properties of the Kubernetes API wrapping the application Helm chart are the fields defined in the chart’s values.yaml file. Application deployments can be customized by specifying different values for these spec properties.

KubePlus architecture details are available [here](http://kubeplus-docs.s3-website-us-west-2.amazonaws.com/html/index.html).
KubePlus is a referenced solution for [multi-customer tenancy in Kubernetes](https://kubernetes.io/docs/concepts/security/multi-tenancy/#multi-customer-tenancy).


## Example

Let’s look at an example of creating a multi-instance WordPress Service using KubePlus. The WordPress service provider goes through the following steps towards this:

1) Set the Namespace in which to deploy KubePlus

   ``export KUBEPLUS_NS=<namespace in which you want to run KubePlus>``

2) Create provider kubeconfig using the provider-kubeconfig.py utility that we provide

   ``python provider-kubeconfig.py create $KUBEPLUS_NS``

3) Install KubePlus Operator using the generated provider kubeconfig 

   ```
   helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-3.0.11.tgz?raw=true" --kubeconfig=kubeplus-saas-provider.json -n $KUBEPLUS_NS
   until kubectl get pods -A | grep kubeplus | grep Running; do echo "Waiting for KubePlus to start.."; sleep 1; done
   ```

4) Create API wrapping WordPress Helm chart.
   
   The WordPress Helm chart can be specified as a [public url](./examples/multitenancy/wordpress/wordpress-service-composition.yaml#L14) or can be [available locally](./examples/multitenancy/wordpress/wordpress-service-composition-localchart.yaml#L14).

   ```
   kubectl create -f ./examples/multitenancy/wordpress/wordpress-service-composition.yaml --kubeconfig=kubeplus-saas-provider.json
   until kubectl get crds --kubeconfig=kubeplus-saas-provider.json | grep wordpressservices  ; do echo "Waiting for WordPressService CRD to be registered.."; sleep 1; done
   ```

5) Create WordpressService instance1

   ``kubectl create -f ./examples/multitenancy/wordpress/tenant1.yaml --kubeconfig=kubeplus-saas-provider.json``

6) Create WordpressService instance2

   ``kubectl create -f ./examples/multitenancy/wordpress/tenant2.yaml --kubeconfig=kubeplus-saas-provider.json``

7) Check created WordpressService instances

   ``kubectl get wordpressservices``

   ```
   NAME             AGE
   wp-for-tenant1   86s
   wp-for-tenant2   26s
   ```

8) Check created application resources

   ``kubectl appresources WordpressService wp-for-tenant1 –k kubeplus-saas-provider.json``

    ```
     NAMESPACE                 KIND                      NAME                      
     default                   WordpressService          wp-for-tenant1            
     wp-for-tenant1            PersistentVolumeClaim     mysql-pv-claim            
     wp-for-tenant1            PersistentVolumeClaim     wp-for-tenant1            
     wp-for-tenant1            Service                   wordpress-mysql           
     wp-for-tenant1            Service                   wp-for-tenant1            
     wp-for-tenant1            Deployment                mysql                     
     wp-for-tenant1            Deployment                wp-for-tenant1            
     wp-for-tenant1            Pod                       mysql-76d6d9bdfd-7ffhx    
     wp-for-tenant1            Pod                       wp-for-tenant1-87c4c954-7n9rk 
     wp-for-tenant1            NetworkPolicy             allow-external-traffic    
     wp-for-tenant1            NetworkPolicy             restrict-cross-ns-traffic 
     wp-for-tenant1            ResourceQuota             wordpressservice-wp-for-tenant1 
    ```
<!--
<p align="center">
<img src="./docs/app-resources.png" width="700" height="250" class="center">
</p>-->

9) Check application resource consumption

   ``kubectl metrics WordpressService wp-for-tenant1 $KUBEPLUS_NS -k kubeplus-saas-provider.json``
   
   ```
   ---------------------------------------------------------- 
   Kubernetes Resources created:
       Number of Sub-resources: -
       Number of Pods: 2
           Number of Containers: 2
           Number of Nodes: 1
           Number of Not Running Pods: 0
   Underlying Physical Resoures consumed:
       Total CPU(cores): 0.773497m
       Total MEMORY(bytes): 516.30859375Mi
       Total Storage(bytes): 40Gi
       Total Network bytes received: 0
       Total Network bytes transferred: 0
   ---------------------------------------------------------- 
   ```

10) Cleanup

    ```
    kubectl delete -f ./examples/multitenancy/wordpress/tenant1.yaml --kubeconfig=kubeplus-saas-provider.json
    kubectl delete -f ./examples/multitenancy/wordpress/tenant2.yaml --kubeconfig=kubeplus-saas-provider.json
    kubectl delete -f ./examples/multitenancy/wordpress/wordpress-service-composition.yaml --kubeconfig=kubeplus-saas-provider.json
    helm delete kubeplus -n $KUBEPLUS_NS
    python provider-kubeconfig.py delete $KUBEPLUS_NS
    ```

<!--
<p align="center">
<img src="./docs/app-metrics.png" width="700" height="250" class="center">
</p>-->

## Try

1) Create minikube cluster

   ``$ minikube start --kubernetes-version=v1.24.3``

2) Download KubePlus plugins and set up the PATH
```
  wget https://github.com/cloud-ark/kubeplus/blob/master/kubeplus-kubectl-plugins.tar.gz
  gunzip kubeplus-kubectl-plugins.tar.gz
  tar -xvf kubeplus-kubectl-plugins
  export KUBEPLUS_HOME=`pwd`
  export PATH=$KUBEPLUS_HOME/plugins:$PATH
```

3) Try following examples:
   - [Hello world](./examples/multitenancy/hello-world/steps.txt)
   - [Wordpress](./examples/multitenancy/wordpress/steps.txt)
   - [Odoo](./examples/multitenancy/odoo/steps.txt)

4) Run tests
```
   cd tests
   python3 -m unittest -v tests
```


## CNCF Landscape

KubePlus is part of CNCF landscape's
[Application Definition section](https://landscape.cncf.io/card-mode?category=application-definition-image-build&grouping=category).


## Operator Maturity Model

As enterprise teams build their custom Kubernetes platforms using community or in house developed Operators, they need a set of guidelines for Operator readin
ess in multi-Operator and multi-tenant environments.
We have developed the [Operator Maturity Model](https://github.com/cloud-ark/kubeplus/blob/master/Guidelines.md) for this purpose. Operator developers are using this model today to ensure that their Operator is a good citizen of the multi-Operator world and ready
 to serve multi-tenant workloads. It is also being used by Kubernetes cluster administrators for curating community Operators towards building their custom pl
atforms.


## Presentations

1. [DevOps.com Webinar: Deliver your Kubernetes Applications as-a-Service](https://webinars.devops.com/deliver-your-kubernetes-applications-as-a-service)

2. [Being a good citizen of the Multi-Operator world, Kubecon NA 2020](https://www.youtube.com/watch?v=NEGs0GMJbCw&t=2s)

3. [Operators and Helm: It takes two to Tango, Helm Summit 2019](https://youtu.be/F_Dgz1V5Q2g)

4. [KubePlus presentation at community meetings (CNCF sig-app-delivery, Kubernetes sig-apps, Helm)](https://github.com/cloud-ark/kubeplus/blob/master/KubePlus-presentation.pdf)


## Contact

For support and new features [reach out to us](https://cloudark.io/kubeplus-saas-manager)
or contact our team on [Slack](https://join.slack.com/t/cloudark/shared_invite/zt-2yp5o32u-sOq4ub21TvO_kYgY9ZfFfw).
