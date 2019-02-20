# TESTING STEPS

`$ eval $(minikube docker-env)`

`$ ./build.sh`

`$ kubectl create -f testpod.yaml`

`$ kubectl logs etcd-helper etcd-tester`
