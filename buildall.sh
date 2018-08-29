#!/bin/bash

echo "Building Operator Manager"
rm operator-manager/operator-manager
docker rmi operator-manager
cd operator-manager
dep ensure
./build-local-deploy-artifacts.sh
cd ..

echo "Building Operator Deployer"
rm operator-deployer/operator-deployer
docker rmi operator-deployer
cd operator-deployer
dep ensure
./build-local-deploy-artifacts.sh
cd ..

echo "Building Kube discovery APIServer"
docker rmi kube-discovery-apiserver
cd ../kubediscovery
dep ensure
./build-discovery-artifacts.sh
cd ..
