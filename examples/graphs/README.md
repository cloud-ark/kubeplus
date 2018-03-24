Kubernetes Resource Relationships
==================================

The Kubernetes resource relationship graphs available in this folder are generated
using KubePlus ``kubectl connections`` plugin.

KubePlus kubectl plugins enable discovery, monitoring and troubleshooting of a Kubernetes cluster. You can install them following these steps

Note: Make sure that your Kubernetes version is <= v1.20.0.

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   $ gunzip kubeplus-kubectl-plugins.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

KubePlus's ``kubectl connections`` plugin enables discovering Kubernetes resource relationship graphs. You can use it with any Kubernetes resource (built-in resources like Pod, Deployment, or custom resources like MysqlCluster, Jenkins, etc.).
Here is how you can use the ``kubectl connections`` plugin:

```
NAME
        kubectl connections

SYNOPSIS
        kubectl connections <Kind> <Instance> <Namespace> [-k <Absolute path to kubeconfig>] [-o json|png|flat|html] [-i <Kind1:Instance1,Kind1:Instance1>] [-n <label|specproperty|envvariable|annotation>]

DESCRIPTION
        kubectl connections shows how the input resource is connected to other Kubernetes resources through one of the following 
        types of relationships: ownerReference, labels, annotations, spec property.
OPTIONS
        kubectl connections takes following optional flags as input.
        -k <Absolute path to kubeconfig file>
        -o <json|png|flat>
            This flag controls what type of output to generate.
        -i <Kind1:Instance1,Kind2:Instance2>
            This flag defines which Kinds and instances to ignore when traversing the resource graph.
            kubectl connections will not discover the sub-graphs starting at such nodes.
        -n <label|specproperty|envvariable|annotation>
            This flag defines the relationship types whose details should not be displayed in the graphical output (png).
            You can specify multiple values as comma separated list.
```