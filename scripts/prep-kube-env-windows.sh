#!/bin/bash

mv /vagrant/.kube/config /vagrant/.kube/config.orig
sed 's/C:.*\.minikube/\/vagrant\/.minikube/'g /vagrant/.kube/config.orig | sed 's/\\/\//'g > /vagrant/.kube/config
cp -r /vagrant/.kube ~/.
#wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
#gunzip kubeplus-kubectl-plugins.tar.gz
#tar -xvf kubeplus-kubectl-plugins.tar
#export KUBEPLUS_HOME=`pwd`
#export PATH=$KUBEPLUS_HOME/plugins/:$PATH
#kubectl kubeplus commands 
#exec

