==================
Operator Manager 
==================

This is a Kubernetes Operator for managing other Operators.

The goal of this Operator is to support deployment of Operators from their Helm Charts.

This Operator is implemented as Kubernetes Custom Resource Definition (CRD). 


How does it work?
=================

A new 'kind' named 'Operator' is defined (see artifacts/deployment/deployment.yaml).

The Custom Resource Controller (controller.go) listens for the creation of resources
of the 'Operator' kind (e.g.: artifacts/examples/test-operator.yaml).
In the spec of a Operator resource you can define:

- Name that you want to give to this Operator

- CharURL: URL of the Helm Chart for the Operator that you want to install

The controller handles Operator resource creation event by triggering Helm deployment.

How to test?
============

Pre-requisite step:
-------------------
1) Install Go's dep dependency management tool:
   https://github.com/golang/dep


Conceptual Steps:
------------------

One time steps:

- Run the Operator Manager

Steps that will be run multiple times for multiple customers:

- Create Operator custom resources


Actual steps:
--------------
0) If using Minikube, enable using local docker images:
 
   - eval $(minikube docker-env)

1) Clone this repository and put it inside 'src' directory of your GOPATH
   at following location:

   $GOPATH/src/github.com/cloud-ark/kubeplus

2) Install dependencies:

   - cd $GOPATH/src/github.com/cloud-ark/kubeplus/operator-manager

   - dep ensure

3) Build local deploy artifacts:

   - ./build-local-deploy-artifacts.sh

4) Deploy Operator Manager

   - kubectl apply -f artifacts/deployment/deployment-minikube.yaml

5) Deploy an Operator

   - kubectl apply -f artifacts/examples/test-operator.yaml

6) Verify
 
   - kubectl get pods

   - Select the name of the Pod corresponding to operator-manager

   - kubectl logs <pod-name-from-above-step>


Suggestions/Issues:
====================

Suggestions to improve this CRD are welcome. Please submit a Pull request, or
give your suggestions here:

https://github.com/cloud-ark/kubeplus/issues

