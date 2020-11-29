#!/bin/bash

kubectl create -f artifacts/deployment/deployment.yaml
kubectl create -f artifacts/examples/resourcecomposition.yaml
sleep 10 

kubectl create -f artifacts/examples/mystack1.yaml
