# Contributing to KubePlus

We would love your contributions. The process is simple:

## Code submissions:

Set up your development environment by following the steps [here](https://github.com/cloud-ark/kubeplus/wiki/Development-and-Testing).

When submitting a PR, link to the [relevant GitHub issue](https://github.com/cloud-ark/kubeplus/issues).

Ping us on KubePlus Slack channel for any questions/discussions about the feature/issue that you are working on.


## Feature requests:

File an Issue with the following information:

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
    - export KUBEPLUS_NS=<Namespace in which KubePlus is deployed>
    - kubectl logs <kubeplus-pod> -n $KUBEPLUS_NS -c crd-hook
    - kubectl logs <kubeplus-pod> -n $KUBEPLUS_NS -c helmer
    - kubectl logs <kubeplus-pod> -n $KUBEPLUS_NS -c platform-operator
    - kubectl logs <kubeplus-pod> -n $KUBEPLUS_NS -c webhook-cert-setup
    - kubectl logs <kubeplus-pod> -n $KUBEPLUS_NS -c consumerui
    - kubectl exec -it <kubeplus-pod> -n $KUBELUS_NS -c kubeconfiggenerator /bin/bash; tail -100 /root/kubeconfiggenerator.log
  ```

  * Cluster details

    * Kubernetes version

    * If using minikube
    
      * minikube version

    * If using Cloud K8S:
  
      * Cloud provider, VM configuration (OS, CPU, RAM, Disk)
