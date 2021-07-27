===================================
Setting up development environment
===================================

In order to develop KubePlus you need following tools:
``Golang (1.14.5), Python 3, Docker (latest), Minikube (latest), kubectl (latest), helm (latest).``

We provide a Vagrantfile that installs all these tools as part of provisioning the Vagrant VM. Using Vagrant is not a requirement for setting up local development environment though. You can also install these tools locally if that is more convenient for you.

Code setup
------------

- Install Vagrant (latest)
- Install VirtualBox (latest)

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
	$ export PATH=$PATH:/usr/local/go/bin
	$ go version

ADD PATH=$PATH:/usr/local/go/bin to ~/.profile

.. code-block:: bash

    $ vi ~/.profile
	$ source ~/.profile

In case any of the above commands fail, you can manually install the tools inside
the Vagrant VM. Here is the command for installing Golang.

.. code-block:: bash

	$ wget -c https://dl.google.com/go/go1.14.5.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local

Note that the ``kubeplus`` folder on your host machine is mapped under ``/vagrant``
directory inside the Vagrant VM. Any files that you want to copy back from the Vagrant VM to the host, place them in ``/vagrant`` folder. Then access them from your host machine in the ``kubeplus`` folder.


Test sample examples
---------------------

Start the Minikube Kubernetes cluster.

.. code-block:: bash

	$ minikube start --driver=docker
	$ kubectl get pods -A

Once Kubernetes cluster is up, try the ``hello-world`` example by following steps from `getting started guide`_.

.. _getting started guide: https://cloud-ark.github.io/kubeplus/docs/html/html/getting-started.html


Work on the code
-----------------

In order to work on the code, you can clone a fresh copy of the code and place it in the path where Golang compiler expects it (which is, ``~/go/src/github.com/cloud-ark``).

.. code-block:: bash

	$ mkdir -p go/src/github.com/cloud-ark
	$ cd go/src/github.com/cloud-ark/
	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus

If you don't want to re-clone kubeplus then create a symbolic link from ``/vagrant`` to
``~/go/src/github.com/cloud-ark``. As noted above, the ``/vagrant`` folder from inside your Vagrant VM is the mapped ``kubeplus`` folder on the host.


Platform Operator
------------------

.. code-block:: bash

	$ cd platform-operator
	$ export GO111MODULE=off
	$ go build .
	$ cd ..


Helm Pod
---------

.. code-block:: bash

	$ cd platform-operator/helm-pod
	$ export GO111MODULE=on
	$ go get github.com/googleapis/gnostic@v0.4.0
	$ go build .
	$ cd ../../


Mutating Webhook
-----------------

.. code-block:: bash

	$ cd mutating-webhook
	$ go build .
	$ cd ..


Mutating Webhook Helper
------------------------

.. code-block:: bash

	$ cd mutating-webhook-helper
	$ go build .
	$ cd ..



