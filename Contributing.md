# Contributing to KubePlus

We would love your contributions. The process is simple:


## Feature requests:

File an Issue with following information:

  * Feature description

  * The context in which the feature need arose


## Bug reports:

File an Issue with following information:

  * KubePlus Helm chart version.

  * Sample YAMLs that you were using:
    - ResourceComposition YAML definition
    - Service YAMLs

  * KubePlus logs

  ```
    - export KUBEPLUS_NAMESPACE=<Namespace in which KubePlus is deployed>
    - kubectl logs $KUBEPLUS_NAMESPACE -c crd-hook
    - kubectl logs $KUBEPLUS_NAMESPACE -c helmer
    - kubectl logs $KUBEPLUS_NAMESAPCE -c platform-operator
    - kubectl logs $KUBEPLUS_NAMESAPCE -c webhook-cert-setup
    - kubectl logs $KUBEPLUS_NAMESAPCE -c consumerui
  ```

  * Cluster details

    * Kubernetes version

    * If using minikube
    
      * minikube version

    * If using Cloud K8S:
  
      * Cloud provider, VM configuration (OS, CPU, RAM, Disk)
