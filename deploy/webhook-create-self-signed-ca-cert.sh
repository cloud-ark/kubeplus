#!/bin/bash -x 
# original source of file (istio dev):
# https://github.com/morvencao/kube-mutating-webhook-tutorial/blob/master/deployment/webhook-create-signed-cert.sh
set -e

usage() {
    cat <<EOF
Generate certificate suitable for use with an sidecar-injector webhook service.

This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with sidecar-injector webhook
services. This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explantion and additional instructions.

The server key/cert k8s CA cert are stored in a k8s secret.

usage: ${0} [OPTIONS]

The following flags are required.

       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
       --secret           Secret name for CA certificate and server certificate/key pair.
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --secret)
            secret="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

echo $namespace
echo $secret
echo $service
csrName=${service}.${namespace}

# Check if mutatingwebhookconfiguration object is present; if so, we can assume that the webhook has been installed;
op=$(kubectl get mutatingwebhookconfigurations platform-as-code.crd-binding 2>&1 || true) 
if [[ $op == *"AGE"* ]]; then
   echo "Mutating webhook is already configured."
   exit
fi

# Source: https://www.funkypenguin.co.nz/blog/self-signed-certificate-on-mutating-webhook-requires-double-encryption/
# 1. Create CA key and CA cert
openssl genrsa -out rootCA.key 4096
openssl req -x509 -new -nodes -key rootCA.key -days 35600 -out rootCA.crt -subj "/CN=admission_ca"

cat <<EOF >> csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

# 2. Create server key, server csr, sign the server csr, and save the signed server cert
openssl genrsa -out server.key 2048 
openssl req -new -key server.key -subj "/CN=${service}.${namespace}.svc" -out server.csr -config csr.conf
openssl x509 -req -in server.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out serverCrt -days 365 -extensions v3_req -extfile csr.conf
openssl x509 -noout -text -in ./serverCrt

# 3. create a secret with signed server cert and server key
kubectl delete secret ${secret} -n ${namespace} 2>/dev/null || true
kubectl create secret generic ${secret} \
        --from-file=key.pem=server.key \
        --from-file=cert.pem=serverCrt \
        --dry-run=client -o yaml |
    kubectl -n ${namespace} apply -f -

# 4. create the mutatingwebhookconfiguration object
sed -i s"/namespace:.*/namespace: $namespace/"g /root/mutatingwebhook.yaml
abc=$(cat rootCA.crt | base64 )
export CA_BUNDLE=$(echo $abc | sed 's/ //'g)
cat /root/mutatingwebhook.yaml | sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" > /root/mutatingwebhook-ca-bundle.yaml
more /root/mutatingwebhook-ca-bundle.yaml

kubectl apply -f ./root/mutatingwebhook-ca-bundle.yaml 2>/dev/null || true

#sleep 10

#kubectl delete -f /root/kubeplus-non-pod-resources.yaml 2>/dev/null || true
#kubectl create -f /root/kubeplus-non-pod-resources.yaml 2>/dev/null || true

#python3 /root/kubeconfiggenerator.py $namespace

#sleep 3

#kubectl label --overwrite=true ns $namespace managedby=kubeplus

