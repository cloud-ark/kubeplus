#!/bin/bash -x

# original source of file (istio dev):
# https://github.com/morvencao/kube-mutating-webhook-tutorial/blob/master/deployment/webhook-patch-ca-bundle.sh
ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail

#kubectl config view --raw -o json | sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[0].cluster.certificateauthdata'

#export CA_BUNDLE=$(/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | select(.name == "'$(/root/kubectl config current-context)'") | .cluster.certificateauthdata')

#/root/kubectl config view --raw --flatten -o json
#/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g 

#export CA_BUNDLE=$(/root/kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | .cluster.certificateauthdata')

#export CA_BUNDLE=$(/root/kubectl get secret -o jsonpath="{.items[?(@.type==\"kubernetes.io/service-account-token\")].data['ca\.crt']}")
#echo $CA_BUNDLE
export CA_BUNDLE=$(kubectl get secrets | grep default | awk '{print $1'} | xargs kubectl get secret -o jsonpath="{.data['ca\.crt']}")

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi
