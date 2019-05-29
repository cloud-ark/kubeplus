Referring to: https://itnext.io/automated-tls-with-cert-manager-and-letsencrypt-for-kubernetes-7daaa5e0cae4

# building docker image:
$ `cd images`

$ `docker build -t example-nodejs:1.0 ./`

$ `docker tag example-nodejs:1.0 gcr.io/kubernetes-dev-211403/example-nodejs:1.0`

$ `docker push gcr.io/kubernetes-dev-211403/example-nodejs`


# status
I have referenced the moodle.go createService, createDeployment, createIngress, and adjusted the app.yaml accordingly. The port has to be 30000-32000 because it is nodeport, and it looks like in moodle.go it does use the same MOODLE_PORT for everything. But I am not able to connect to or curl my kube cluster, with ip http://34.66.187.39:30080/.

`kubectl create -f app.yaml`

`kubectl get ingress` `kubectl get deployment` `kubectl get service`

next step is to get the domain name tls stuff to work with the domain I registered, then will need to add a couple annotations to moodle.go
