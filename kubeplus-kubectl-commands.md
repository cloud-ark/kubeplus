KubePlus kubectl plugins
-------------------------

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

Available commands:

```
     NAME
             kubectl kubeplus commands
     
     SYNOPSIS
             kubectl man
             kubectl connections
             kubectl metrics
             kubectl applogs
             kubectl retrieve kubeconfig provider
             kubectl retrieve kubeconfig consumer
     
     DESCRIPTION
             KubePlus provides a suite of kubectl plugins to discover, monitor and troubleshoot Kubernetes applications.
     
             The discovery plugins (kubectl man and kubectl connections) help with discovering the static and runtime
             information about an application.
             - kubectl man provides the ability to discover man page like information about Kubernetes Custom Resources.
             - kubectl connections provides the ability to discover Kubernetes resources that are related to one another
               through one of the following relationships - ownerReferences, label, annotations, spec properties.
             The monitoring and troubleshooting plugins (kubectl metrics and kubectl applogs) enable collecting application metrics and logs.
             - kubectl metrics collects CPU, Memory, Storage, and Network metrics for an application. These are available in Prometheus format.
             - kubectl applogs collects logs for all the containers of all the Pods in an application.
             The kubeconfig files that are meant to be used by SaaS provider and SaaS consumers are available through:
             - kubectl retrieve kubeconfig provider
             - kubectl retrieve kubeconfig consumer
             These kubeconfig files are provided with limited RBAC permissions appropriate for the persona.
```