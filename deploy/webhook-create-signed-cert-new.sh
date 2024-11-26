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
tmpdir=$(mktemp -d)
echo "creating certs in tmpdir ${tmpdir} "

cat <<EOF >> ${tmpdir}/csr.conf
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

openssl genrsa -out ${tmpdir}/server-key.pem 2048
openssl req -new -key ${tmpdir}/server-key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf

# clean-up any previously created CSR for our service. Ignore errors if not present.
#./root/kubectl delete csr ${csrName} 2>/dev/null || true
./root/kubectl delete csr ${csrName} 2>/dev/null || true

# create  server cert/key CSR and  send to k8s API
cat <<EOF | ./root/kubectl create --validate=false -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${csrName}
spec:
  groups:
  - system:authenticated
  request: $(cat ${tmpdir}/server.csr | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

# verify CSR has been created
while true; do
    ./root/kubectl get csr ${csrName}
    if [ "$?" -eq 0 ]; then
        break
    fi
done

# approve and fetch the signed certificate
./root/kubectl certificate approve ${csrName}
# verify certificate has been signed
for x in $(seq 10); do
    serverCert=$(./root/kubectl get csr ${csrName} -o jsonpath='{.status.certificate}')
    if [[ ${serverCert} != '' ]]; then
        break
    fi
    sleep 1
done
if [[ ${serverCert} == '' ]]; then
    echo "ERROR: After approving csr ${csrName}, the signed certificate did not appear on the resource. Giving up after 10 attempts." >&2
    exit 1
fi
echo ${serverCert} | openssl base64 -d -A -out ${tmpdir}/server-cert.pem


# create the secret with CA cert and server cert/key
./root/kubectl delete secret ${secret} -n ${namespace} 2>/dev/null || true
./root/kubectl create secret generic ${secret} \
        --from-file=key.pem=${tmpdir}/server-key.pem \
        --from-file=cert.pem=${tmpdir}/server-cert.pem \
        --dry-run=client -o yaml |
    ./root/kubectl -n ${namespace} apply -f -

sed -i s"/namespace:.*/namespace: $namespace/"g /root/mutatingwebhook.yaml
cat ./root/mutatingwebhook.yaml | ./root/webhook-patch-ca-bundle-new.sh $namespace > ./root/mutatingwebhook-ca-bundle.yaml
more /root/mutatingwebhook-ca-bundle.yaml

kubectl apply -f ./root/mutatingwebhook-ca-bundle.yaml 2>/dev/null || true

kubectl delete -f /root/kubeplus-non-pod-resources.yaml 2>/dev/null || true
kubectl create -f /root/kubeplus-non-pod-resources.yaml 2>/dev/null || true

python3 /root/kubeconfiggenerator.py $namespace

kubectl label --overwrite=true ns $namespace managedby=kubeplus
