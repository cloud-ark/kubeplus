#!/bin/bash

IMAGE_NAME="custom-hello-world-app:latest"

eval $(minikube docker-env)
docker build -t "$IMAGE_NAME" $KUBEPLUS_HOME/examples/managed-service/appmetrics/custom-hello-world/
