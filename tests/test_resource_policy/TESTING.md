# ResourcePolicy network-policy test

This walkthrough tests a standalone ResourcePolicy. It assumes KubePlus is
installed in `default`, including the ResourcePolicy CRD, and uses an updated
platform-operator image.

Run the walkthrough from the KubePlus repository root directory.

Start the test profile with Calico, which enforces NetworkPolicies, and select
its context before running the commands below:

```bash
minikube start -p kubeplus-netpol --cni=calico
kubectl config use-context kubeplus-netpol
```

## Optional local image update

```bash
eval "$(minikube docker-env -p kubeplus-netpol)"

cd platform-operator
GOFLAGS=-mod=readonly ./build-artifact.sh latest
cd ..

kubectl -n default set image deployment/kubeplus-deployment \
  platform-operator=gcr.io/cloudark-kubeplus/platform-operator:latest
kubectl -n default rollout restart deployment/kubeplus-deployment
kubectl -n default rollout status deployment/kubeplus-deployment --timeout=180s
```

## Clean start

```bash
kubectl -n default delete resourcepolicy rag-access --ignore-not-found
kubectl delete namespace team-b --ignore-not-found --wait=true
kubectl delete namespace rag-services team-a --ignore-not-found --wait=true
```

## Create and inspect

```bash
cd tests/test_resource_policy

kubectl apply -f rag-network-test.yaml
kubectl -n rag-services rollout status deployment/rag-api --timeout=120s
kubectl -n team-a wait --for=condition=Ready pod/chatbot --timeout=120s

kubectl apply -f ResourcePolicy.yaml
kubectl -n default get resourcepolicy rag-access

for i in $(seq 1 30); do
  kubectl -n rag-services get networkpolicy chatbot-to-rag >/dev/null 2>&1 && break
  sleep 1
done
kubectl -n rag-services get networkpolicy chatbot-to-rag -o yaml
```

The generated policy should target `partof=ragservice-shared-rag`, allow
`team-a`/`partof=chatbot-chatbot`, use TCP port `80`, and contain
`kubeplus.io/resource-policy-ref: default/rag-access`.

## Connectivity

```bash
kubectl -n team-a exec chatbot -- \
  curl -fsS --max-time 5 \
  http://rag-api.rag-services.svc.cluster.local/
```

## Update

```bash
kubectl -n default patch resourcepolicy rag-access --type=json \
  -p='[{"op":"replace","path":"/spec/policy/network/access/0/ports/0/port","value":8080}]'

kubectl -n rag-services wait \
  --for=jsonpath='{.spec.ingress[0].ports[0].port}'=8080 \
  networkpolicy/chatbot-to-rag --timeout=30s
kubectl -n rag-services get networkpolicy chatbot-to-rag \
  -o jsonpath='{.spec.ingress[0].ports[0].port}{"\n"}'
```

The output should be `8080`. Restore the fixture value:

```bash
kubectl -n default patch resourcepolicy rag-access --type=json \
  -p='[{"op":"replace","path":"/spec/policy/network/access/0/ports/0/port","value":80}]'
kubectl -n rag-services wait \
  --for=jsonpath='{.spec.ingress[0].ports[0].port}'=80 \
  networkpolicy/chatbot-to-rag --timeout=30s
```

## Delete and recreate

```bash
kubectl -n default delete resourcepolicy rag-access

for i in $(seq 1 30); do
  kubectl -n rag-services get networkpolicy chatbot-to-rag >/dev/null 2>&1 || break
  sleep 1
done
kubectl -n rag-services get networkpolicy chatbot-to-rag --ignore-not-found

kubectl apply -f ResourcePolicy.yaml
kubectl -n rag-services wait --for=create \
  networkpolicy/chatbot-to-rag --timeout=30s
kubectl -n rag-services get networkpolicy chatbot-to-rag -o yaml
```

For plugin CRUD checks, add the local plugin directory to `PATH` and replace
the native `kubectl apply/get/delete resourcepolicy` commands with:

```bash
export PATH="$(pwd)/../../plugins:$PATH"
kubectl resource policy apply -f ResourcePolicy.yaml
kubectl resource policy get rag-access -n default -o yaml
kubectl resource policy delete rag-access -n default
```

If reconciliation fails, inspect the controller:

```bash
kubectl -n default logs deployment/kubeplus-deployment \
  -c platform-operator --since=5m
```
