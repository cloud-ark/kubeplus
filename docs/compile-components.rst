===================================
Setting up development environment
===================================

In order to develop KubePlus you need following tools:
``Golang (1.14.5), Python 3, Docker (latest), Minikube (latest), kubectl (latest), helm (latest).``

We provide a Vagrantfile that installs all these tools as part of provisioning the Vagrant VM. You can also install these tools locally if that is more convenient for you.

If you want to use Vagrant based environment, follow these steps.

- Install Vagrant (latest)
- Install VirtualBox (latest)
- Install Git Bash (for Windows)

If you are on Windows host then open a git bash terminal and perform
the following steps through that terminal.
Note that adding the box and spinning up the Vagrant VM can take some
time for the first time. On Windows hosts we have noticed that the
git bash terminal can become stuck. In such a case open another git bash
terminal and ssh into the Vagrant VM. 

.. code-block:: bash

	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus
	$ vagrant box add bento/ubuntu-18.04
	$ vagrant up

Once Vagrant VM has started

.. code-block:: bash

	$ vagrant ssh
	$ sudo usermod -aG docker $USER
	$ exit
	$ vagrant ssh
	$ docker ps
	$ docker version
	$ minikube version
	$ kubectl version
	$ helm version

In case any of the above commands fail, manually install the tools inside the Vagrant VM. 

Note that the ``kubeplus`` folder on your host machine is mapped under ``/vagrant`` directory inside the Vagrant VM. Any files that you want to copy back from the Vagrant VM to the host, place them in ``/vagrant`` folder. Then access them from your host machine in the ``kubeplus`` folder.


Test sample examples
---------------------

Start the Minikube Kubernetes cluster.

.. code-block:: bash

	$ minikube start --driver=docker
	$ kubectl get pods -A

Once Kubernetes cluster is up, try the ``hello-world`` example by following steps from `getting started guide`_.

.. _getting started guide: https://cloud-ark.github.io/kubeplus/docs/html/html/getting-started.html


Vagrant VM access
------------------

- Vagrant VM IP: ``192.168.33.10``

- Access consumer ui on Vagrant VM (example of HelloWorldService): ``http://192.168.33.10:5000/service/HelloWorldService#``

- SSH into Vagrant VM: ``vagrant ssh``

- Copy files from Vagrant VM to the Host. Example

.. code-block:: bash

	$ kubectl connections Pod kubeplus-deployment-fddd-ddd default -o png
	$ Output available in: /home/vagrant/plugins/connections-op.json.gv.png
	$ cp plugins/connections-op.json.gv.png /vagrant/.

On the Host go to the directory where you have cloned kubeplus. The copied
file will be available there.


Work on the code
-----------------

Make sure that Golang is installed correctly on the Vagrant VM.

.. code-block:: bash

	$ export PATH=$PATH:/usr/local/go/bin
	$ go version

ADD PATH=$PATH:/usr/local/go/bin to ~/.profile

.. code-block:: bash

	$ vi ~/.profile
	$ source ~/.profile

In case Golang is not properly installed, here is the command to install it.

.. code-block:: bash

	$ wget -c https://dl.google.com/go/go1.14.5.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local

In order to work on the code, you can clone a fresh copy kubeplus and place it in the path where Golang compiler expects it (which is, ``~/go/src/github.com/cloud-ark``).

.. code-block:: bash

	$ cd ~/
	$ mkdir -p go/src/github.com/cloud-ark
	$ cd go/src/github.com/cloud-ark/
	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus
	$ export KUBEPLUS_HOME=/home/vagrant/go/src/github.com/cloud-ark/kubeplus/

If you don't want to re-clone kubeplus then create a symbolic link from ``/vagrant`` to
``~/go/src/github.com/cloud-ark``. As noted above, the ``/vagrant`` folder from inside your Vagrant VM is the mapped ``kubeplus`` folder on the host.

Download gnostic library separately. It is a dependency of one of the k8s projects, but it has been removed from the googleapis project. kubeplus build fails as it depends on several k8s projects. We go around this issue by downloading it separately.

.. code-block:: bash

	$ go get github.com/googleapis/gnostic@v0.4.0

Connect the Docker cli in the VM to the Docker daemon that is part of Minikube.
We need to do this to use the locally built images when testing code changes.

.. code-block:: bash

	$ eval $(minikube docker-env)

Now we are ready to work on the code.


Code Organization
------------------

KubePlus is made up of following components:

- an init container that sets up required KubePlus artifacts such as ServiceAccounts, CRDs, etc. (available in ``$KUBEPLUS_HOME/deploy/`` folder)
- the mutating webhook (available in ``$KUBEPLUS_HOME/mutating-webhook`` folder)
- a mutating webhook helper (available in ``$KUBEPLUS_HOME/mutating-webhook-helper`` folder)
- the platform operator (available in ``$KUBEPLUS_HOME/platform-operator`` folder)
- the helmer container (available in ``$KUBEPLUS_HOME/platform-operator/helm-pod`` folder)
- consumerui (available in ``$KUBEPLUS_HOME/consumerui`` folder)


Use vi/emacs to modify any part of the code.
In order to test the changes, you will need to build the code, deploy KubePlus, 
and run some example (``hello-world`` is a good example for testing purposes).


Build code
-----------
In each of the above component folders a build script is provided (``./build-artifact.sh``). Use it as follows to build the code:

.. code-block:: bash

	$ ./build-artifact.sh latest

Deploy KubePlus
----------------

.. code-block:: bash

	$ cd $KUBEPLUS_HOME/deploy
	$ kubectl create -f kubeplus-components-minikube.yaml

The ``kubeplus-components-minikube.yaml`` refers to the latest tags for each of the components. Also, the ``imagePullPolicy`` is set to ``Never``. If you want to test a particular component tag available on CloudARK's public GCR then don't forget to change the imagePullPolicy to either ``IfNotPresent`` or ``Always``.

Build Failure
--------------

If you see ``ErrImageNeverPull`` or ``CrashLoopBackOff`` then it means that you have not compiled all the components mentioned above. Go to each component directory and compile each component. Then delete KubePlus deployment and try again.


Check Logs
-----------

.. code-block:: bash

	$ cd $KUBEPLUS_HOME/deploy
	$ ./kubeplus-logs.sh

Delete KubePlus
----------------

.. code-block:: bash

	$ cd $KUBEPLUS_HOME/deploy
	$ ./delete-kubeplus-components.sh 


Following components are written in Golang. If you run into any issues with building them then use the following commands to separately try the build steps to debug the issue. 


**Platform Operator**

.. code-block:: bash

	$ cd platform-operator
	$ ./build-artifact.sh latest
	$ export GO111MODULE=off
	$ go build .
	$ cd ..


**Helm Pod**

.. code-block:: bash

	$ cd platform-operator/helm-pod
	$ export GO111MODULE=on
	$ go build .
	$ cd ../../


**Mutating Webhook**

.. code-block:: bash

	$ cd mutating-webhook
	$ export GO111MODULE=on
	$ go build .
	$ cd ..


**Mutating Webhook Helper**

.. code-block:: bash

	$ cd mutating-webhook-helper
	$ export GO111MODULE=on
	$ go build .
	$ cd ..



