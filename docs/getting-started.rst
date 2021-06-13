=================
Getting Started
=================

Setup
------

Install Helm v3 and install KubePlus using following command.
KubePlus can be installed in any Namespace. 

.. code-block:: bash

    $ KUBEPLUS_NS=default (or any namespace in which you want to install KubePlus)
    $ helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-0.2.0.tgz?raw=true" -n $KUBEPLUS_NS

Examples
---------

1. Try `hello world service`_

.. _hello world service: https://cloud-ark.github.io/kubeplus/docs/html/html/sample-example.html


2. Try example outlined in Kubeplus Components section by following steps `here`_.

.. _here: https://github.com/cloud-ark/kubeplus/blob/master/examples/resource-composition/steps.txt

3. Other SaaS examples:

  - `Wordpress service`_
  - `Mysql service`_
  - `MongoDB service`_
  - Multiple `teams with applications deployed later`_

.. _Wordpress service: https://github.com/cloud-ark/kubeplus/blob/master//examples/multitenancy/wordpress-mysqlcluster-stack/steps.txt

.. _Mysql service: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/stacks/steps.txt

.. _MongoDB service: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/mongodb-as-a-service/steps.md

.. _teams with applications deployed later: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/team/steps.txt

4. Build your own SaaS:
   
   - Install Helm version 3.0+
   - Create Helm chart for your application stack and make it available at a publicly accessible URL
   - Follow steps similar to above examples

5. Debug:

.. code-block:: bash

    $ KUBEPLUS=`kubectl get pods -A | grep kubeplus | awk '{print $2}'`
    $ KUBEPLUS_NS=`kubectl get pods -A | grep kubeplus | awk '{print $1}'`
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c crd-hook
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c helmer
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c platform-operator
    $ kubectl logs $KUBEPLUS -n $KUBEPLUS_NS -c webhook-cert-setup

6. Cleanup:

.. code-block:: bash

    $ wget https://github.com/cloud-ark/kubeplus/raw/master/deploy/delete-kubeplus-components.sh
    $ ./delete-kubeplus-components.sh
