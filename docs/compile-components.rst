===================================
Setting up development environment
===================================

Below instructions are given for setting up KubePlus development environment.
These should work for Linux and Mac OS hosts.

Install Go version 1.14.5

.. code-block:: bash

	$ wget -c https://dl.google.com/go/go1.14.5.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
	$ export PATH=$PATH:/usr/local/go/bin
	$ ADD PATH=$PATH:/usr/local/go/bin to ~/.profile
	$ source ~/.profile
	$ go version


Code setup
------------

.. code-block:: bash

	$ mkdir -p go/src/github.com/cloud-ark
	$ cd go/src/github.com/cloud-ark/
	$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
	$ cd kubeplus


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


Local testing using Vagrant
----------------------------

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
	$ minikube start --driver=docker
	$ kubectl get pods -A

Once Kubernetes cluster is up, follow steps from `getting started guide`_

.. _getting started guide: https://cloud-ark.github.io/kubeplus/docs/html/html/getting-started.html

Note that the ``kubeplus`` folder on your host machine is mapped under ``/vagrant``
directory inside the Vagrant VM. Any files that you want to copy back from the Vagrant VM to the host, place them in ``/vagrant`` folder. Then access them from your host machine in the ``kubeplus`` folder.








