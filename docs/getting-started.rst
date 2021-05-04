========================
Getting Started
========================

1. Install KubePlus kubectl plugins:

.. code-block:: bash

  $ Install Docker
  $ 
	$ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   	$ gunzip kubeplus-kubectl-plugins.tar.gz
   	$ tar -xvf kubeplus-kubectl-plugins.tar
   	$ export KUBEPLUS_HOME=`pwd`
   	$ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   	$ kubectl kubeplus commands
    $ kubectl connections ServiceAccount default default -o png

``kubectl connections`` can be used with any Kubernetes resource (built-in resources like Pod, Deployment, or custom resources like MysqlCluster).

2. Install KubePlus in-cluster component before trying out below examples:

.. code-block:: bash

	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus/deploy
	$ ./deploy-kubeplus.sh

We also provide a Helm chart (v3) (available inside kubeplus/deploy directory)
- Install Helm version 3

.. code-block:: bash

  $ helm install kubeplus kubeplus-chart --set caBundle=$(kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | select(.name == "'$(kubectl config current-context)'") | .cluster.certificateauthdata')


3. CRD for CRDs:

   - Try example outlined in Kubeplus Components section by following steps `here`_.

.. _here: https://github.com/cloud-ark/kubeplus/blob/master/examples/resource-composition/steps.txt

4. SaaS examples:

  - `Helloworld service`_
  - `Wordpress service`_
  - `Mysql service`_
  - `MongoDB service`_
  - Multiple `teams with applications deployed later`_

.. _Helloworld service: https://github.com/cloud-ark/kubeplus/blob/master//examples/multitenancy/hello-world/steps.txt

.. _Wordpress service: https://github.com/cloud-ark/kubeplus/blob/master//examples/multitenancy/wordpress-mysqlcluster-stack/steps.txt

.. _Mysql service: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/stacks/steps.txt

.. _MongoDB service: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/mongodb-as-a-service/steps.md

.. _teams with applications deployed later: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/team/steps.txt

5. Try with your own Helm chart:
   
   - Install Helm version 3.0+
   - Create Helm chart and make it available at a publicly accessible URL
   - Follow steps similar to above examples

6. Debug:

  - kubectl logs kubeplus -c crd-hook
  - kubectl logs kubeplus -c helmer
  - kubectl logs kubeplus -c platform-operator
  - kubectl logs kubeplus -c webhook-cert-setup


7. Contributing:
   
   We would love your contributions. The process is simple_.

.. _simple: https://github.com/cloud-ark/kubeplus/blob/master/Contributing.md


OpenShift Market Place Deployment
-----------------------------------

1. Install KubePlus Pre-requisite resources

.. code-block:: bash

    kubectl create -f https://raw.githubusercontent.com/cloud-ark/kubeplus/master/deploy/kubeplus-openshift-prereqs.yaml

2. Install Metrics API Server

.. code-block:: bash

    kubectl create -f https://raw.githubusercontent.com/cloud-ark/kubeplus/master/deploy/metrics-server.yaml

3. Install KubePlus SaaS Manager
   - Follow the standard steps for installing an Operator on OpenShift

4. Try out KubePlus kubectl plugins
    See above

5. Try `Helloworld service`_
