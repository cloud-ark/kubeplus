====================
Discovery Manager
====================

Component of KubePlus that enables discovering static and dynamic information about Custom Resources.

This directory is just a place-holder directory. Actual Discovery Manager code is in kubediscovery_ repository.

.. _kubediscovery: https://github.com/cloud-ark/kubediscovery


Discovery Manager is a Kubernetes Aggregated API Server that registers two custom endpoints: /explain and /composition. You can pass query parameters to these to find out information about Custom Resources.

Example:

Find out usage details about a Custom Resource:

``   $ kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"``

Find out dynamic composition tree for Postgres custom resource instance:

``   $ kubectl get --raw "/apis/platform-as-code/v1/composition?kind=Moodle&instance=moodle1&namespace=namespace1" | python -mjson.tool``
