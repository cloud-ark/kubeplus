Compiling different components:
-------------------------------

Install Go version 1.14.5

```
$ wget -c https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
$ export PATH=$PATH:/usr/local/go/bin
$ ADD PATH=$PATH:/usr/local/go/bin to ~/.profile
$ source ~/.profile
$ go version
```

Code setup:
------------

```
$ mkdir -p go/src/github.com/cloud-ark
$ cd go/src/github.com/cloud-ark/
$ git clone --depth 1 https://github.com/cloud-ark/kubeplus.git
$ cd kubeplus
```

Platform Operator
------------------

```
$ cd platform-operator
$ GO111MODULE=off
$ go build .
$ cd ..
```

Helm Pod
---------

```
$ cd platform-operator/helm-pod
$ GO111MODULE=on
$ go get -u k8s.io/client-go@v0.17.2 github.com/googleapis/gnostic@v0.3.1 ./...
$ go build .
$ cd ../../
```

Mutating Webhook
-----------------

```
$ cd mutating-webhook
$ go build .
$ cd ..
```

Mutating Webhook Helper
------------------------

```
$ cd mutating-webhook-helper
$ go build .
$ cd ..
```



