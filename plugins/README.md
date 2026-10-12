## KubePlus kubectl plugins

Checkout [KubePlus repository](https://github.com/cloud-ark/kubeplus) for details.


## Use it:

```
   $ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins-latest.tar.gz
   $ gunzip kubeplus-kubectl-plugins-latest.tar.gz
   $ tar -xvf kubeplus-kubectl-plugins-latest.tar
   $ export KUBEPLUS_HOME=`pwd`
   $ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   $ kubectl kubeplus commands
```

ResourcePolicy CRUD is available through the thin `kubectl resource-policy`
plugin, which delegates to native `kubectl` commands:

```
$ kubectl resource-policy apply -f ResourcePolicy.yaml
$ kubectl resource-policy get -A
$ kubectl resource-policy edit rag-access -n default
$ kubectl resource-policy delete rag-access -n default
```
