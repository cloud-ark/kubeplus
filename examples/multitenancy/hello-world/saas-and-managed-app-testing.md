SaaS and Managed Application delivery example
==============================================

KubePlus Operator can be set up in two ways - multi-namespace and single namespace.
The multi-namespace configuration is ideal when a separate instance of an application
needs to be created in a separate namespace. The single namespace configuration is
ideal for managed application scenario where an application provider is 
delivering their application on someone else's cluster using KubePlus.

The difference between the two configurations is that, for multi-namespace configuration, the KubePlus Operator needs to be installed in the
default Namespace whereas for single namespace configuration, the KubePlus Operator
needs to be installed in any other namespace. When installed in the default Namespace,
KubePlus gets permissions to create new namespaces and deploy applications inside them.

Use Helm version 3+. With minikube, you can create a cluster with a specific version like so:
```
    $ minikube start --kubernetes-version=v1.24.3
```

Setup KubePlus kubectl plugins
-------------------------------
```
$ wget https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz
$ gunzip kubeplus-kubectl-plugins.tar.gz
$ tar -xvf kubeplus-kubectl-plugins.tar
$ export KUBEPLUS_HOME=`pwd`
$ export PATH=$KUBEPLUS_HOME/plugins/:$PATH
$ kubectl kubeplus commands
```



Multi-namespace setup
----------------------
```
1. KUBEPLUS_NS=default
2. helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-3.0.0.tgz?raw=true" -n $KUBEPLUS_NS
3. kubectl create ns testns --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus
   - request should be allowed
4. kubectl create ns testns1 --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-provider
   - request should be denied
5. kubectl create ns testns2 --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-consumer
   - request should be denied
6.  until kubectl get pods -n $KUBEPLUS_NS | grep Running; do echo "Waiting for KubePlus Operator to become ready"; sleep 1; done
7. kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-consumer\.json}" > consumer.conf
8.  more consumer.conf | grep namespace
    - namespace should be set to 'default'
9. kubectl create -f hello-world-service-composition.yaml --kubeconfig=consumer.conf
   - request should be denied
10. kubectl get configmaps kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-provider\.json}" > provider.conf
11. more provider.conf | grep namespace
    - namespace should be set to 'default'
12. kubectl create -f hello-world-service-composition.yaml --kubeconfig=provider.conf
13. until kubectl get crds --kubeconfig=provider.conf | grep hello  ; do echo "Waiting for HelloworldService CRD to be registered.."; sleep 1; done
14. kubectl man HelloWorldService -k consumer.conf
15. kubectl create -f hs1.yaml --kubeconfig=consumer.conf
16. kubectl get helloworldservices hs1 -o json
17. kubectl get pods -A
    - Hello World Pod in hs1 namespace
18. kubectl get ns
    - hs1 namespace has been created
19. kubectl appurl HelloWorldService hs1 default -k consumer.conf
    - curl the IP address received. Should see "Hello hello hello"
20. kubectl applogs HelloWorldService hs1 default -k consumer.conf
    - Should see app logs
21. kubectl connections HelloWorldService hs1 default -i Namespace:$KUBEPLUS_NS -k consumer.conf
    - Should see created resources' listing
22. kubectl metrics HelloWorldService hs1 -o prometheus -k consumer.conf
    - Should see the metrics
23. kubectl delete -f hs1.yaml --kubeconfig=consumer.conf
    - kubectl get pods -A
      - Hello World Pod is deleted
24. kubectl get ns
    - hs1 namespace has been deleted
25. kubectl delete ns testns --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus
    - should be allowed
26. helm delete kubeplus -n $KUBEPLUS_NS
```

Single namespace setup
-----------------------
```
1. kubectl create ns kubeplus
2. KUBEPLUS_NS=kubeplus
3. helm install kubeplus "https://github.com/cloud-ark/operatorcharts/blob/master/kubeplus-chart-3.0.0.tgz?raw=true" -n $KUBEPLUS_NS
4. kubectl create ns testns --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus
   - request should be denied
5. kubectl create ns testns1 --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-provider
   - request should be denied
6. kubectl create ns testns2 --as=system:serviceaccount:$KUBEPLUS_NS:kubeplus-saas-consumer
   - request should be denied
7.  until kubectl get pods -n $KUBEPLUS_NS | grep Running; do echo "Waiting for KubePlus Operator to become ready"; sleep 1; done
8. kubectl get configmaps kubeplus-saas-consumer-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-consumer\.json}" > consumer.conf
9.  more consumer.conf | grep namespace
    - namespace should be set to 'kubeplus'
10. kubectl create -f hello-world-service-composition.yaml --kubeconfig=consumer.conf
    - request should be denied
11. kubectl get configmaps kubeplus-saas-provider-kubeconfig -n $KUBEPLUS_NS -o jsonpath="{.data.kubeplus-saas-provider\.json}" > provider.conf
12. more provider.conf | grep namespace
    - namespace should be set to 'kubeplus'
13. kubectl create -f hello-world-service-composition.yaml --kubeconfig=provider.conf
14. until kubectl get crds --kubeconfig=provider.conf | grep hello  ; do echo "Waiting for HelloworldService CRD to be registered.."; sleep 1; done
15. kubectl man HelloWorldService -k provider.conf
16. kubectl create -f hs1.yaml --kubeconfig=provider.conf
17. kubectl get pods -A
    - Hello World Pod in kubeplus namespace
18. kubectl appurl HelloWorldService hs1 kubeplus -k provider.conf
    - curl the IP address received. Should see "Hello hello hello"
19. kubectl applogs HelloWorldService hs1 kubeplus -k provider.conf
    - Should see app logs
20. kubectl connections HelloWorldService hs1 kubeplus -i Namespace:$KUBEPLUS_NS -k provider.conf
    - Should see created resources' listing
21. kubectl metrics HelloWorldService hs1 -o prometheus -k provider.conf
    - Should see the metrics
22. kubectl delete -f hs1.yaml --kubeconfig=provider.conf
23. kubectl get pods -A
    - Hello World Pod is deleted
24. kubectl get ns
    - kubeplus namespace remains intact
25. helm delete kubeplus -n $KUBEPLUS_NS
26. kubectl delete ns kubeplus
```

KubePlus Operator troubleshooting:
-----------------------------------
If any of the above steps fail (don't give expected output), collect the logs and share them with us in a Github issue.

```
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c crd-hook
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c helmer
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c platform-operator
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c webhook-cert-setup
  - kubectl logs <kubeplus-pod> $KUBEPLUS_NS -c consumerui
```

