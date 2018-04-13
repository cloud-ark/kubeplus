=========================
Postgres Custom Resource 
=========================

This is a Kubernetes Custom Resource for Postgres.

The goal of this Custom Resource Definition (CRD) is to support various life-cycle actions 
for a Postgres instance, such as:

- Create user and password when Postgres instance is created
- Create database when Postgres instance is created
- Modify user password on an existing Postgres instance
- Modify user permissions on an existing Postgres instance
- etc.

This CRD reduces out-of-band automation that you have to implement for provisioning
a Postgres instance on Kubernetes and managing its various life-cycle actions.


How it works?
=============

A new 'kind' named 'Postgres' is defined (see artifacts/examples/crd.yaml).

The Custom Resource Controller (controller.go) listens for the creation of resources
of the 'Postgres' kind (e.g.: artifacts/examples/client1-postgres.yaml).
In the spec of a Postgres resource you can define any commands that you want to execute
on the provisioned Postgres instance. 

The controller handles Postgres resource creation event by creating a 
Kubernetes Deployment with the specified Postgres image in the CRD definition, 
and creating a Kubernetes Service to expose this Deployment. 
Currently the created Service is of type NodePort as it makes it easy to test
the controller on Minikube. In real deployments this can be changed to LoadBalancer
type of Service. It is also possible to use an Ingress resource to expose the
Service at some path.

Once the Postgres Pod is READY, the controller executes commands defined in the 
commands spec attribute of the custom resource against the Service endpoint.

An example custom resource is shown below.

--------------
kind: Postgres
metadata:
  name: client1
spec:
  deploymentName: client1
  username: client1
  password: client1
  database: moodle
  image: postgres:9.3
  replicas: 1
  commands: ["create user client1 with password 'client1';","create database moodle with owner client1 encoding 'utf8' template=template0;"]
----------------

The commands attribute contains any commands that you want to execute on the Postgres instance.
(TODO: The commands definition should be changed to accommodate parameterized variable. This will allow
using username, password, and database name defined as separate attributes in the spec to be integrated/used
within the commands definition.)

This and other example Postgres custom resources are available in following directory:
./artifact/examples

controller.go: This file contains the logic of handling the custom resource creation.


How to test?
============

The code has been developed and tested on Minikube. 

Minikube's IP address is hard coded in controller.go
- Find out the IP address of Minikube VM: minikube ip
- Update the MINIKUB_IP variable defined in controller.go


Pre-requisite step:
-------------------
1) Install Go

2) Set GOPATH to point to folder where you will maintain src, bin, pkg folders
   for any go code. One option is to use $HOME/go. 
   This folder is Workspace folder for your Go code
   - mkdir -p $HOME/go/src $HOME/go/bin $HOME/go/pkg
   - export GOPATH=$HOME/go
   - export GO_WORKSPACE=$GOPATH/src

3) Add the bin folder of Go installation and the bin folder under GOPATH to
   your PATH environment variable:
   GO_INSTALL_BIN=`which go`
   GO_WORKSPACE_BIN=$GOPATH/bin
   export PATH=$PATH:$GO_INSTALL_BIN:$GO_WORKSPACE_BIN

4) Install Go's dep dependency management tool:
   https://github.com/golang/dep

5) Install kubectl

6) Install Minikube

7) Install Postgres client:
   - brew install postgresql
   - sudo apt-get install postgresql-client


Conceptual Steps:
------------------
One time steps:
- Run CustomResource Controller for Postgres
- Register the Postgres CRD with the Cluster

Steps that will be run multiple times for multiple customers:
- Create Postgres custom resource


Actual steps (Minikube):
-------------------------
1) Start Minikube VM
   - minikube start

2) Clone this repository:
   - git clone git@github.com:cloud-ark/kubeplus.git

3) Symlink the kubeplus folder into your Go Workspace folder at
   appropriate location:
   - cd $GO_WORKSPACE
   - mkdir -p github.com/cloud-ark ; cd github.com/cloud-ark
   - ln -s <path-where-you-cloned-kubeplus> kubeplus

4) Install dependencies:
   - cd $GO_WORKSPACE/github.com/cloud-ark/kubeplus
   - dep ensure

5) In one shell window run Postgres custom resource controller
   - cd $GO_WORKSPACE/github.com/cloud-ark/kubeplus/postgres-crd
   - go run *.go -kubeconfig=$HOME/.kube/config

6) In another shell window register CRD definition for Postgres
   - export GOPATH=$HOME/go
   - export GO_WORKSPACE=$GOPATH/src
   - cd $GO_WORKSPACE/github.com/cloud-ark/kubeplus/postgres-crd
   - kubectl create -f artifacts/examples/crd.yaml
   - kubectl get crd

7) In the second window create Postgres custom resource for client1
   - kubectl create -f artifacts/examples/client1-postgres.yaml 

  
Verify:
--------
1) kubectl get crd

2) kubectl get postgres client1

3) kubectl describe postgres client1

4) minikube service <service name> --url
   - Parse VM IP and Service Port from the URL

5) psql -h <IP> -p <port> -U <username> -d <db-name>
   - When prompted for password, enter <password>
   - IP: Minikube IP
   - port: Port of the exposed Service
   - username: Name of the user from setupCommands artifacts/examples/client1-postgres.yaml 
   - db-name: Name of the database from setupCommands artifacts/examples/client1-postgres.yaml 
   - password: Value of password from setupCommands artifacts/examples/client1-postgres.yaml 


Suggestions/Issues:
====================

Suggestions to improve this CRD are welcome. Please submit a Pull request, or
give your suggestions here:

https://github.com/cloud-ark/kubeplus/issues

