========================
Getting Started
========================

1. Try ``kubectl connections`` plugin in your environment. It can be used with any Kubernetes resource (built-in resources like Pod, Deployment, or custom resources like MysqlCluster).

.. code-block:: bash

	$ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
   	$ gunzip kubeplus-kubectl-plugins.tar.gz
   	$ tar -xvf kubeplus-kubectl-plugins.tar
   	$ export KUBEPLUS_HOME=`pwd`
   	$ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
   	$ kubectl kubeplus commands

2. Install KubePlus server-side component before trying out below examples:

.. code-block:: bash

	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus/deploy
	$ ./deploy-kubeplus.sh

3. CRD for CRDs:

   - Try example outlined in Kubeplus Components section by following steps `here`_.

.. _here: https://github.com/cloud-ark/kubeplus/blob/master/examples/resource-composition/steps.txt

4. Multitenancy examples:

   - Try following multitenancy examples:
     - Multiple `application stacks`_
     - Multiple `teams with applications deployed later`_

.. _application stacks: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/stacks/steps.txt

.. _teams with applications deployed later: https://github.com/cloud-ark/kubeplus/blob/master/examples/multitenancy/team/steps.txt

5. Try with your own Helm chart:
   
   - Install Helm version 3.0+
   - Create Helm chart and make it available at a publicly accessible URL
   - Follow steps similar to above examples