===================================
Sample Example - HelloWorldService
===================================

Here we demonstrate how a Provider can use KubePlus to deliver a "hello-world as-a-Service" using a Hello World Helm chart.
The helm chart defines a Deployment and a Service. The Pod managed
by the Deployment prints the messages that are provided as input.
When registering the service, Provider defines the CPU and Memory requests and limits that should be provided to the HelloWorld Pod. KubePlus ensures that every HelloWorldService instance Pod is given the CPU and Memory requests and limits that are configured by the Provider when registering the HelloWorldService. Consumers create HelloWorldService instances and pass custom hello messages to print to the Pod.

Setup
------

In order to try this example you will need to install Kubernetes CLI (kubectl), or if you are using OpenShift, the OpenShift CLI (oc).
Choose the CLI version that works with your Kubernetes version.
Once the appropriate CLI is installed, follow these steps.

Open two command terminal windows. Name them as:

- Provider window
- Consumer window 

KubePlus Namespace
--------------------

Set KUBEPLUS_NS environment variable to the NAMESPACE in which you have installed KubePlus. For OpenShift users, the namespace needs to be 'openshift-operators'.

.. code-block:: bash

    KUBEPLUS_NS=<Namespace>

Get provider and consumer kubeconfigs
--------------------------------------

KubePlus generates separate kubeconfig files for provider and consumers with appropriate permissions. Retrieve them as follows:

.. code-block:: bash

	$ kubectl retrieve kubeconfig provider $KUBEPLUS_NS > provider.conf
	$ kubectl retrieve kubeconfig consumer $KUBEPLUS_NS > consumer.conf

In the steps below, use the appropriate kubeconfig in the provider and consumer actions by passing the ``--kubeconfig=<provider/consumer>.conf`` flag.


Register HelloWorldService (provider window)
-------------------------------------------------

1. Create hello-world-resource-composition:

Here is the hello-world-resource-composition.yaml file. Save it as hello-world-resource-composition.yaml.

.. code-block:: bash

	apiVersion: workflows.kubeplus/v1alpha1
	kind: ResourceComposition
	metadata:
	  name: hello-world-resource-composition
	spec:
	  # newResource defines the new CRD to be installed define a workflow.
	  newResource:
	    resource:
	      kind: HelloWorldService
	      group: platformapi.kubeplus
	      version: v1alpha1
	      plural: helloworldservices
	    # URL of the Helm chart that contains Kubernetes resources that represent a workflow.
	    chartURL: https://github.com/cloud-ark/operatorcharts/blob/master/hello-world-chart-0.0.2.tgz?raw=true
	    chartName: hello-world-chart
	  # respolicy defines the resource policy to be applied to instances of the specified custom resource.
	  respolicy:
	    apiVersion: workflows.kubeplus/v1alpha1
	    kind: ResourcePolicy 
	    metadata:
	      name: hello-world-service-policy
	    spec:
	      resource:
	        kind: HelloWorldService 
	        group: platformapi.kubeplus
	        version: v1alpha1
	      policy:
	        # Add following requests and limits for the first container of all the Pods that are related via 
	        # owner reference relationship to instances of resources specified above.
	        podconfig:
	          limits:
	            cpu: 200m
	            memory: 2Gi
	          requests:
	            cpu: 100m
	            memory: 1Gi
	  # resmonitor identifies the resource instances that should be monitored for CPU/Memory/Storage.
	  # All the Pods that are related to the resource instance through either ownerReference relationship, or all the relationships
	  # (ownerReference, label, annotation, spec properties) are considered in calculating the statistics. 
	  # The generated output is in Prometheus format.
	  resmonitor:
	    apiVersion: workflows.kubeplus/v1alpha1
	    kind: ResourceMonitor
	    metadata:
	      name: hello-world-service-monitor
	    spec:
	      resource:
	        kind: HelloWorldService 
	        group: platformapi.kubeplus
	        version: v1alpha1
	      # This attribute indicates that Pods that are reachable through all the relationships should be used
	      # as part of calculating the monitoring statistics.
	      monitorRelationships: all

The ``respolicy`` section in the resource composition defines the ``ResourcePolicy`` that the provider configures for this service. Here it defines the cpu and memory requests and limits that need to be configured for service instances of this service.  

Create hello-world-resource-composition as follows:

.. code-block:: bash

    kubectl create -f hello-world-resource-composition.yaml -n $KUBEPLUS_NS --kubeconfig=provider.conf

or

.. code-block:: bash

    oc create -f hello-world-resource-composition.yaml -n $KUBEPLUS_NS --kubeconfig=provider.conf


2. Wait till HelloWorldService CRD is registered in the cluster.

.. code-block:: bash

    until kubectl get crds | grep hello  ; do echo "Waiting for HelloworldService CRD to be registered.."; sleep 1; done

or

.. code-block:: bash

    until oc get crds | grep hello  ; do echo "Waiting for HelloworldService CRD to be registered.."; sleep 1; done


3. Grant permission to the consumer to create service instances.

.. code-block:: bash

	kubectl grantpermission consumer helloworldservices provider.conf $KUBEPLUS_NS


Create HelloWorldService instance (consumer window)
----------------------------------------------------

HelloWorldService instances can be created using either kubectl or consumer ui that
KubePlus provides.


**Using Consumer UI**

The consumer UI is part of KubePlus and runs on the cluster. Access it as follows:

.. code-block:: bash

	$ wget https://raw.githubusercontent.com/cloud-ark/kubeplus/master/deploy/open-consumer-ui.sh
	$ chmod +x open-consumer-ui.sh
	$ ./open-consumer-ui.sh

The HelloWorldService will be available at following URL:

.. code-block:: bash

	$ http://localhost:5000/service/HelloWorldService

If you are working with the KubePlus Vagrant VM, access the service at following URL:

.. code-block:: bash

	$ http://192.168.33.10:5000/service/HelloWorldService

The UI provides a form to input values that need to be provided when creating a service instance. You can also check the API documentation for the service on the UI.
Because the provider has granted permission to the consumer to create the HelloWorldService instances, you will be able to create an instance of HelloWorldService through the UI. Once a service instance has been created, the UI displays cpu, memory, storage, network metrics for the instance, and resource relationship graph which shows all the Kubernetes resources that are created as part of that instance and how they are related to one another.

**Using CLI**


1. Install KubePlus kubectl plugins

.. code-block:: bash

    curl -L https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz -o kubeplus-kubectl-plugins.tar.gz
    gunzip kubeplus-kubectl-plugins.tar.gz
    tar -xvf kubeplus-kubectl-plugins.tar
    export KUBEPLUS_HOME=`pwd`
    export PATH=$KUBEPLUS_HOME/plugins/:$PATH
    kubectl kubeplus commands
  or
    oc kubeplus commands

2. Install Docker and verify that you are able to run docker commands without requiring sudo.

.. code-block:: bash

	docker ps

This should return without any errors.

3. Check the HelloWorldService API documentation

.. code-block:: bash

	kubectl man HelloWorldService

You should see following output:

.. code-block:: bash

	KIND:	HelloWorldService
	GROUP:	platformapi.kubeplus
	VERSION:	v1alpha1

	DESCRIPTION:
	Here is the values.yaml for the underlying Helm chart representing this resource.
	The attributes in values.yaml become the Spec properties of the resource.

	::::::::::::::
	/hello-world-chart/values.yaml
	::::::::::::::
	# Default value for namespace.

	greeting: Hello World!


4. Create HelloWorldService instance:

Copy below YAML and save it as hello-world-service.yaml

.. code-block:: bash

	apiVersion: platformapi.kubeplus/v1alpha1
	kind: HelloWorldService 
	metadata:
	  name: hs1
	spec:
	  greeting: Hello hello hello

.. code-block:: bash

    kubectl create -f hello-world-service.yaml --kubeconfig=consumer.conf

or

.. code-block:: bash

    oc create -f hello-world-service.yaml --kubeconfig=consumer.conf

This will create hs1 instance in the default namespace.


5. Check if the service instance has been created:

.. code-block:: bash

    kubectl get helloworldservices --kubeconfig=consumer.conf
    kubectl describe helloworldservices hs1 --kubeconfig=consumer.conf

or

.. code-block:: bash

    oc get helloworldservices --kubeconfig=consumer.conf
    oc describe helloworldservices hs1 --kubeconfig=consumer.conf

Verify that the Status field is populated in hs1 instance.


6. Verify that HelloWorldService has been started

.. code-block:: bash

    HELLOWORLD_POD=`kubectl get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $2}'`
    HELLOWORLD_NS=`kubectl get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $1}'`
    kubectl port-forward $HELLOWORLD_POD -n $HELLOWORLD_NS 8082:5000 &
    curl localhost:8082

or

.. code-block:: bash

    HELLOWORLD_POD=`oc get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $2}'`
    HELLOWORLD_NS=`oc get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $1}'`
    oc port-forward $HELLOWORLD_POD -n $HELLOWORLD_NS 8082:5000 &
    curl localhost:8082

You should see following output:

.. code-block:: bash

	Hello hello hello

7. Verify resource requests and limits have been set on the Pod that belongs to HelloWorldService instance.

.. code-block:: bash

	kubectl get pods $HELLOWORLD_POD -n $HELLOWORLD_NS -o json | jq -r '.spec.containers[0].resources'

or

.. code-block:: bash
   
    oc get pods $HELLOWORLD_POD -n $HELLOWORLD_NS -o json | jq -r '.spec.containers[0].resources'


You should see following output:

.. image:: hello-world-resources.png
   :align: center
   :height: 150px
   :width: 200px

8. Check resource relationship graph for HelloWorldService instance:

.. code-block:: bash

    kubectl connections HelloWorldService hs1 $HELLOWORLD_NS

or

.. code-block:: bash

    oc connections HelloWorldService hs1 $HELLOWORLD_NS

You should see following output:

.. image:: hello-world-connections-flat.png
   :align: center

9. (Only on Linux or MacOS) Visualize the relationship graph:

.. code-block:: bash

    kubectl connections HelloWorldService hs1 $HELLOWORLD_NS -o png

or

.. code-block:: bash

    oc connections HelloWorldService hs1 $HELLOWORLD_NS -o png


.. image:: hello-world-connections-png.png
   :align: center


Get HelloWorldService instance metrics (provider/consumer window)
---------------------------------------------------------------------

On the provider window or the consumer window, perform following steps:

.. code-block:: bash

    KUBEPLUS_POD=`kubectl get pods -A | grep kubeplus-deployment | awk '{print $2}'`

    KUBEPLUS_NS=`kubectl get pods -A | grep kubeplus-deployment | awk '{print $1}'`

    kubectl port-forward $KUBEPLUS_POD -n $KUBEPLUS_NS 8081:8090 &

or

.. code-block:: bash

    KUBEPLUS_POD=`oc get pods -A | grep kubeplus-deployment | awk '{print $2}'`

    KUBEPLUS_NS=`oc get pods -A | grep kubeplus-deployment | awk '{print $1}'`

    oc port-forward $KUBEPLUS_POD -n $KUBEPLUS_NS 8081:8090 &


Get cpu, memory, storage, network metrics for HelloWorldService instance:

.. code-block:: bash

    HELLOWORLD_NS=`kubectl get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $1}'`

or

.. code-block:: bash

    HELLOWORLD_NS=`oc get pods -A | grep hello-world-deployment-helloworldservice | awk '{print $1}'`


.. code-block:: bash

    curl -kv "http://127.0.0.1:8081/apis/kubeplus/metrics?kind=HelloWorldService&instance=hs1&namespace=$HELLOWORLD_NS"

You should see output of the following form:

.. image:: hello-world-metrics.png
   :align: center
