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



