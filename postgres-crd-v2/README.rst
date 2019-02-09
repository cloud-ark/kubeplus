==================
Postgres Operator
==================

This is a Kubernetes Operator for Postgres.

The goal of this Operator is to support various life-cycle actions 
for a Postgres instance, such as:

- Create user and password when Postgres instance is created
- Create database at the time of Postgres instance creation
- Initialize the database with some data
- Create database on an already created Postgres instance
- Modify user password on an existing Postgres instance
- etc.

This Operator is implemented as Kubernetes Custom Resource Definition (CRD). 
It reduces out-of-band automation that you have to implement for provisioning
a Postgres instance on Kubernetes and managing its various life-cycle actions.


How does it work?
=================

A new 'kind' named 'Postgres' is defined (see artifacts/examples/crd.yaml).

The Operator registers Postgres CRD at start up (in main.go).

The Custom Resource Controller (controller.go) listens for the creation of resources
of the 'Postgres' kind (e.g.: artifacts/examples/initializeclient.yaml).
In the spec of a Postgres resource you can define 

- Databases that you want created using the 'database' attribute
- Users with their passwords that you want created using the 'users' attribute
- The 'initcommands' attribute should be used to specify any table creation and
  data insert commands. See artifacts/examples/initializeclient.yaml for example.

The controller handles Postgres resource creation event by creating a 
Kubernetes Deployment with the Postgres image specified in the CRD definition.
It exposes this Deployment using a Kubernetes Service.
Currently the created Service is of type NodePort as it makes it easy to test
the controller on Minikube. In real deployments this can be changed to LoadBalancer
type of Service. It is also possible to use an Ingress resource to expose the
Service at some path instead of at an IP address.

The Deployment should be changed to a Stateful Set in real deployments.


How to test?
============

Pre-requisite step:
-------------------
1) Install Go's dep dependency management tool:
   https://github.com/golang/dep

2) Install Postgres client:

- brew install postgresql

- sudo apt-get install postgresql-client


Conceptual Steps:
------------------

One time steps:

- Run the Postgres Operator

Steps that will be run multiple times for multiple customers:

- Create Postgres custom resources


Actual steps:
--------------
0) If using Minikube, enable using local docker images:
 
   - eval $(minikube docker-env)

1) Clone this repository and put it inside 'src' directory of your GOPATH
   at following location:

   $GOPATH/src/github.com/cloud-ark/kubeplus

2) Install dependencies:

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/postgres-crd-v2

   - dep ensure

3) In one shell window run the Postgres Operator

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/postgres-crd-v2

   - Follow either one of the following 2 options:

     A] Deploy the controller as a Deployment in the cluster using
        controller Docker image built locally
     
         - ./build-local-deploy-artifacts.sh
     
         - cd artifacts/deployment

	 - kubectl apply -f default-sa.yaml

         - kubectl create -f deployment-minikube.yaml

     B] Deploy the controller with Helm chart (here the controller
        Docker image is pulled from Docker hub)

        - helm install ./postgres-crd-v2-chart

4) In another shell window check that Postgres CRD has been registered.

   - kubectl get customresourcedefinition

   If it is not registered you can manually register it as follows:

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/postgres-crd-v2

   - kubectl create -f artifacts/examples/crd.yaml


5) In the second window create Postgres custom resource

   - kubectl apply -f artifacts/examples/initializeclient.yaml

   - kubectl describe postgres client25

   - Verify (see below)

6) Test Life-cycle actions (execute and verify)

   - kubectl apply -f artifacts/examples/add-user.yaml

   - kubectl apply -f artifacts/examples/delete-user.yaml 

   - kubectl apply -f artifacts/examples/modify-password.yaml

   - kubectl apply -f artifacts/examples/add-db.yaml

   - kubectl apply -f artifacts/examples/delete-db.yaml

7) Clean up

   - kubectl get deployments

   - kubectl delete deployments ...

   - helm list

   - helm delete ...

   - ./deletecrds.sh ...

   
Verify:
--------
1) kubectl get crd

2) kubectl get postgres client25

3) kubectl describe postgres client25

4) psql -h <IP of the Host> -p <port> -U <username> -d <db-name>
   - When prompted for password, enter <password>
   - IP: Minikube IP (find using 'minikube ip' command)
   - port: Port of the exposed Service
   - username: Name of the user from artifacts/examples/initializeclient.yaml
   - db-name: Name of the database from setupCommands artifacts/examples/initializeclient.yaml
   - password: Value of password from setupCommands artifacts/examples/initializeclient.yaml


Suggestions/Issues:
====================

Suggestions to improve this CRD are welcome. Please submit a Pull request, or
give your suggestions here:

https://github.com/cloud-ark/kubeplus/issues

