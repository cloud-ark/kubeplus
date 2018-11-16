====================
Discovery Manager
====================

Component of KubePlus that enables discovering static and dynamic information about Custom Resources.

This directory is just a place-holder directory. Actual Discovery Manager code is in kubediscovery_ repository.

.. _kubediscovery: https://github.com/cloud-ark/kubediscovery


Discovery Manager is a Kubernetes Aggregated API Server that registers two custom endpoints: /explain and /composition. You can pass query parameters to these to find out information about Custom Resources.

Example:

Find out details about a Custom Resource's Spec definition:

``$ kubectl get --raw "/apis/kubediscovery.cloudark.io/v1/explain?kind=Postgres"``


Find out dynamic composition tree for Postgres custom resource instance:

``$ kubectl get --raw "/apis/kubediscovery.cloudark.io/v1/composition?kind=Postgres&instance=postgres1" | python -mjson.tool``


**Note:**

The interface of ``kubectl get --raw`` for explain will not be needed once upstream ``kubectl explain`` 
starts supporting custom resources.

We plan to contribute kubediscovery_ code upstream towards this.
