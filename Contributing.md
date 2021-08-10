# Contributing to KubePlus

We would love your contributions. The process is simple:


## Feature requests:

File an Issue with following information:

  * Feature description

  * The context in which the feature need arose


## Bug reports:

File an Issue with following information:

  * Kubernetes version

    * kubectl version

  * Helm version

    * helm version

  * Host details

    * If using minikube
    
      * minikube version

    * If using Cloud VM
  
      * Cloud provider, VM configuration (OS, CPU, RAM, Disk)

  * Error log output

    * Include container logs:
  ```
    - kubectl logs kubeplus -c crd-hook
    - kubectl logs kubeplus -c helmer
    - kubectl logs kubeplus -c platform-operator
    - kubectl logs kubeplus -c webhook-cert-setup
  ```

  * Sample YAMLs that you were using

  * KubePlus deployment details: 
    - How was KubePlus deployed? (Helm chart or directly from KubePlus deployment manifests available in deploy folder)
    - If using Helm then the Helm chart version used.
    - If directly from KubePlus deployment manifests then Image tags for all the 
      KubePlus components
