#!/bin/bash

apt-get update
apt-get install -y gcc make socat git wget emacs
wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz 
sudo tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz 
export PATH=$PATH:/usr/local/go/bin 
export GOROOT=/usr/local/go 

mkdir $HOME/goworkspace 
mkdir $HOME/goworkspace/src
mkdir $HOME/goworkspace/bin 

export GOPATH=$HOME/goworkspace

curl -L https://github.com/coreos/etcd/releases/download/v3.2.18/etcd-v3.2.18-linux-amd64.tar.gz -o etcd-v3.2.18-linux-amd64.tar.gz && tar xzvf etcd-v3.2.18-linux-amd64.tar.gz && /bin/cp -f etcd-v3.2.18-linux-amd64/{etcd,etcdctl} /usr/bin && rm -rf etcd-v3.2.18-linux-amd64*

sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common


curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -


sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

sudo apt-get update
sudo apt-get install docker-ce

git clone https://github.com/kubernetes/kubernetes $GOPATH/src/k8s.io/kubernetes 
cd $GOPATH/src/k8s.io/kubernetes 

git checkout remotes/origin/release-1.11

git checkout -b release-1.11

echo "Start cluster by running following"
echo "export KUBERNETES_PROVIDER=local"
echo "nohup hack/local-up-cluster.sh &"