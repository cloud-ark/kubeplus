## KubePlus API Add-on

KubePlus API add-on simplifies building workflow automation using Kubernetes Custom Resoures. It brings uniformity in using different Custom Resources and enables 
usage monitoring and troubleshooting. KubePlus API Add-on is being developed as part of our [Platform as Code practice](https://cloudark.io/platform-as-code)

## kubectl commands

KubePlus API add-on offers following interfaces through kubectl:

**1. kubectl man <Custom Resource>:** Provides information about how to use different  the Custom Resource.

**2. kubectl composition <Custom Resource Instance>:** Provides information about sub resources created by the Operator for handling Custom Resource Instance.

**3. kubectl metrics cr <Custom Resource Instance>:** Provides information about CPU and memory resources consumed by the Custom Resource Instance.

**4. kubectl metrics account <Account Name>:** Provides information about CPU and memory resources consumed by all the Kubernetes resources created by the account.

**5. kubectl crlogs <Custom Resource Instance>:** Provides logs for all the containers for a Custom Resource Instance.

**6. kubectl cr-relations <Custom Resource Instance>: (upcoming)** Provide information about relationships of Custom Resource Instance with other instances via labels / annotations / Spec Properties / sub-resources.

In order to use above kubectl commands all you need to do is enhance your Operator's Custom Resource Definition (CRD) YAML with following two annotations.

```
platform-as-code/usage
```

The 'usage' annotation is used to define usage information for a Custom Resource.
The value for 'usage' annotation is the name of the ConfigMap that stores the usage information.

```
platform-as-code/composition
```

The 'composition' annotation is used to define Kubernetes's built-in resources that are created as part of instantiating a Custom Resource instance. Its value is a list 

As an example, annotations on MysqlCluster Custom Resource Definition (CRD) are shown below:

```
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: mysqlclusters.mysql.presslabs.org
    annotations:
      helm.sh/hook: crd-install
      platform-as-code/usage: mysqlcluster-usage.usage
      platform-as-code/composition: StatefulSet, Service, ConfigMap, Secret, PodDisruptionBudget
  spec:
    group: mysql.presslabs.org
    names:
      kind: MysqlCluster
      plural: mysqlclusters
      shortNames:
      - mysql
    scope: Namespaced
```

Details about KubePlus API Add-on can be found [here](./details.rst).


## Try it:

Use Kubernetes version 1.14

```

   $ git clone https://github.com/cloud-ark/kubeplus.git
   $ cd kubeplus
   $ ./deploy-kubeplus.sh
   $ export PATH=$PATH:`pwd`
```

## Example

``` 
$ kubectl metrics account devdattakulkarni@gmail.com
---------------------------------------------------------- 
 Creator Account Identity: devdattakulkarni@gmail.com
---------------------------------------------------------- 
 Number of Custom Resources: 3
 Number of Deployments: 1
 Number of StatefulSets: 1
 Number of ReplicaSets: 1
 Number of DaemonSets: 1
 Number of ReplicationControllers: 1
 Number of Pods: 1
Total CPU(cores): 259m
Total MEMORY(bytes): 255Mi
```

## Status

Under development

