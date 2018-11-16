# Contributing to KubePlus

We would love your contributions. The process is simple as outlined below:


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

    * kubectl logs <kubeplus-pod-id> -c operator-manager

    * kubectl logs <kubeplus-pod-id> -c operator-deployer

    * kubectl logs <kubeplus-pod-id> -c kube-discovery-apiserver

  * Sample YAMLs that you were using

  * Image tags for operator-manager, operator-deployer, kube-discovery-apiserver 
    from deploy/rc.yaml
