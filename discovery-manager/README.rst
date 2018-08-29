====================
Discovery Manager
====================

Component of KubePlus that enables discovering static and dynamic information about Custom Resources.

Discovery Manager is a Kubernetes Aggregated API Server that registers two custom endpoints: /explain and /describe

You can pass query parameters to these to find out information about Custom Resource.

Example:

Find out details about a Custom Resource's Spec definition:

`$ kubectl get --raw "/apis/kubediscovery.cloudark.io/v1/explain?cr=Postgres"`


Find out dynamic composition tree for Postgres custom resource instance:

`$ kubectl get --raw "/apis/kubediscovery.cloudark.io/v1/describe?cr=Postgres&instance=postgres1" | python -mjson.tool`


Check README from kubeplus repository for detailed example.

This directory is just a place-holder directory. Actual Discovery Manager code is in kubediscovery_ repository.

.. _kubediscovery: https://github.com/cloud-ark/kubediscovery