#!/bin/bash -x

# original source of file (istio dev):
# https://github.com/morvencao/kube-mutating-webhook-tutorial/blob/master/deployment/webhook-patch-ca-bundle.sh
ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail

namespace="default"

if (( $# == 1 )); then
	namespace=$1
fi

#kubectl config view --raw -o json | sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[0].cluster.certificateauthdata'

#export CA_BUNDLE=$(/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | select(.name == "'$(/root/kubectl config current-context)'") | .cluster.certificateauthdata')

#/root/kubectl config view --raw --flatten -o json
#/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g 

#export CA_BUNDLE=$(/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | .cluster.certificateauthdata')

#export CA_BUNDLE=$(/root/kubectl get secret -o jsonpath="{.items[?(@.type==\"kubernetes.io/service-account-token\")].data['ca\.crt']}")
#echo $CA_BUNDLE

### Works - May 21, 2021
export CA_BUNDLE=$(kubectl get csr crd-hook-service.$namespace -o jsonpath='{.status.certificate}')


#kubectl get secrets -n $namespace | grep service-account-token | grep default-token | head -1 | awk '{print $1}' | xargs kubectl get secret -n $namespace -o jsonpath="{.data['ca\.crt']}" | base64 --decode > ca_chain.pem
#csplit -s -z -f individual- ca_chain.pem '/-----BEGIN CERTIFICATE-----/' '{*}'
#base64 -i individual-05 > root-ca.encoded
#export CA_BUNDLE=$(readarray -t ARRAY < root-ca.encoded; IFS=''; echo "${ARRAY[*]}")

#export CA_BUNDLE=`base64 individual-00 | cut -d '\'' -f1`
#export CA_BUNDLE=`base64 individual-00 | sed "s/'//g"`
#export CA_BUNDLE=`cat root-ca.encoded`

# - working export CA_BUNDLE=$(kubectl get secrets -n $namespace | grep service-account-token | head -1 | awk '{print $1}' | xargs kubectl get secret -n $namespace -o jsonpath="{.data['ca\.crt']}")

#export CA_BUNDLE=$(kubectl get configmaps -n $namespace kube-root-ca.crt -o jsonpath="{.data['ca\.crt']}")

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi
