# ResourcePolicy CLI demo

Run from the repository root in Bash. Assume KubePlus is already installed in
`default` on the `kubeplus-netpol` Minikube profile, with Calico enforcing
NetworkPolicies. The test creates two mock HTTP apps.

## Setup

```bash
minikube start -p kubeplus-netpol --driver=docker --cni=calico --keep-context
kubectl config use-context kubeplus-netpol

# Rebuild the operator from the current source and load the image into Minikube.
(cd platform-operator && GOFLAGS=-mod=readonly ./build-artifact.sh latest)
minikube -p kubeplus-netpol image load gcr.io/cloudark-kubeplus/platform-operator:latest --daemon
kubectl apply -f deploy/kubeplus-chart/crds/kubeplus-crds.yaml
kubectl wait --for=condition=Established crd/resourcepolicies.workflows.kubeplus --timeout=60s

# Update the platform-operator container in the existing KubePlus installation.
kubectl -n default set image deployment/kubeplus-deployment \
  platform-operator=gcr.io/cloudark-kubeplus/platform-operator:latest
# Restart existing Pods to pick up the rebuilt image, even though the tag is unchanged.
kubectl -n default rollout restart deployment/kubeplus-deployment
kubectl -n default rollout status deployment/kubeplus-deployment --timeout=180s

export KUBEPLUS_HOME="$PWD"
export PATH="$KUBEPLUS_HOME/plugins:$PATH"
kubectl -n default delete resourcepolicy rag-access cross-namespace-access network-access.chatbot-a.rag-a --ignore-not-found
kubectl apply -f tests/test_resource_policy/mock-apps.yaml
kubectl -n chatbot-a wait --for=condition=Ready pod/chatbot --timeout=120s
kubectl -n rag-a wait --for=condition=Ready pod/rag-api --timeout=120s
```

Define the two requests used after each command:

```bash
probe() {
  kubectl -n chatbot-a exec chatbot -- \
    wget -qO- -T 3 http://rag-api.rag-a.svc.cluster.local:8080/
  kubectl -n rag-a exec rag-api -- \
    wget -qO- -T 3 http://chatbot.chatbot-a.svc.cluster.local:8080/
}
probe
```

Initially both requests should time out. Successful requests print `rag-a` and
then `chatbot-a`. Rule updates are asynchronous; give the controller and Calico a
moment to apply them, then rerun `probe` if necessary.

## Existing Python tool

```bash
kubectl allow network traffic chatbot-a rag-a
probe  # Both succeed.
kubectl deny network traffic chatbot-a rag-a
probe  # Both time out.
```

## ResourcePolicy CLI

```bash
kubectl resource-policy allow-traffic chatbot-a rag-a
kubectl -n default get resourcepolicy network-access.chatbot-a.rag-a -o yaml
probe  # Both succeed.
kubectl resource-policy deny-traffic chatbot-a rag-a
probe  # Both time out.
```

Allow stores one ResourcePolicy for the pair; the controller updates
`restrict-cross-ns-traffic` in both namespaces. Deny deletes that object and the
controller removes its allowances. Local traffic rules remain. Reversed namespace
arguments select the same object. Another ResourcePolicy listing this pair can
still keep traffic allowed, so run the Python comparison first on a clean setup.

Inspect the result or controller logs with:

```bash
kubectl -n chatbot-a get networkpolicy restrict-cross-ns-traffic -o yaml
kubectl -n rag-a get networkpolicy restrict-cross-ns-traffic -o yaml
kubectl -n default logs deployment/kubeplus-deployment -c platform-operator --since=5m
```

The mock setup omits `allow-external-traffic`: other allow policies can mask a
denied pair. To remove the mock apps after the demo:

```bash
kubectl resource-policy deny-traffic chatbot-a rag-a
kubectl delete namespace chatbot-a rag-a
```
