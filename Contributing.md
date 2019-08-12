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

    * kubectl logs <kubeplus-apiserver-pod>

    * kubectl logs <kubeplus-apiserver-pod> -c kube-discovery-apiserver

    * kubectl logs <kubeplus-mutating-webhook-pod>

  * Sample YAMLs that you were using

  * Image tags from deploy/rc.yaml, platform-operator/artifacts/deployment/deployment.yaml,
    mutating-webhook/deployment/deployment.yaml, mutating-webhook-helper/deployment.yaml

