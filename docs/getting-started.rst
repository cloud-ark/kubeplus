=================
Getting Started
=================

Setup
------

Install Helm v3 and install KubePlus using following commands
KubePlus can be installed in any Namespace. 

.. code-block:: bash

    $ wget https://raw.githubusercontent.com/cloud-ark/kubeplus/master/provider-kubeconfig.py
    $ KUBEPLUS_NS=default (or any namespace in which you want to install KubePlus)
    $ python provider-kubeconfig.py create $KUBEPLUS_NS
    $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-3.0.8.tgz?raw=true" --kubeconfig=kubeplus-saas-provider.json -n $KUBEPLUS_NS
    $ until kubectl get pods -A | grep kubeplus | grep Running; do echo "Waiting for KubePlus to start.."; sleep 1; done

Examples
---------

1. `hello world service`_

.. _hello world service: http://kubeplus-docs.s3-website-us-west-2.amazonaws.com/html/sample-example.html



2. `Wordpress service`_

.. _Wordpress service: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/wordpress/steps.txt

3. Build your own SaaS:
   
   - Install Helm version 3.0+
   - Create Helm chart for your application stack and make it available at a publicly accessible URL
   - Follow steps similar to above examples

4. Debug:

.. code-block:: bash

    $ KUBEPLUS=`kubectl get pods -A | grep kubeplus | awk '{print $2}'`
    $ KUBEPLUS_NS=`kubectl get pods -A | grep kubeplus | awk '{print $1}'`
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c crd-hook
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c helmer
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c platform-operator
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c webhook-cert-setup
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c consumerui
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c kubeconfiggenerator
    $ kubectl get configmaps kubeplus-saas-provider -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-provider\.json}" > provider-kubeconfig.json
    $ kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-consumer\.json}" > consumer-kubeconfig.json
    $ kubectl auth can-i --list --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-provider
    $ kubectl auth can-i --list --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-consumer


5. Cleanup:

.. code-block:: bash

    $ helm delete kubeplus -n $KUBEPLUS_NS
    
