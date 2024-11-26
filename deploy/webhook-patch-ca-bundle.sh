#!/bin/bash

# original source of file (istio dev):
# https://github.com/morvencao/kube-mutating-webhook-tutorial/blob/master/deployment/webhook-patch-ca-bundle.sh
ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail

#kubectl config view --raw -o json | sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[0].cluster.certificateauthdata'

export CA_BUNDLE=$(kubectl config view --raw --flatten -o json |  sed 's/certificate-authority-data/certificateauthdata/'g | jq -r '.clusters[] | select(.name == "'$(kubectl config current-context)'") | .cluster.certificateauthdata')

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi
