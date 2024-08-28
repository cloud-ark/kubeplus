#!/bin/bash

# Function to print usage
usage() {
    echo "Usage: $0 [--kubeplus-plugin] [--kubeplus <kubeplus-namespace>]"
    echo "Example: $0 --kubeplus-plugin --kubeplus kubeplus-namespace"
    exit 1
}

# Exit on error
set -e

# Check if Helm is installed
if ! command -v helm &> /dev/null; then
    echo "Error: Helm is not installed. Please install Helm before running this script."
    exit 1
fi

# Check if KUBEPLUS_CI is set and is true
if [ -z "$KUBEPLUS_CI" ]; then
    KUBEPLUS_CI=false
fi

# Parse arguments
TIMEOUT=1200
INSTALL_KUBEPLUS=false
INSTALL_KUBEPLUS_PLUGIN=false
KUBEPLUS_NAMESPACE="kubeplus"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --kubeplus-plugin)
            INSTALL_KUBEPLUS_PLUGIN=true
            ;;
        --kubeplus)
            INSTALL_KUBEPLUS=true
            if [ -n "$2" ] && [[ "$2" != --* ]]; then
                KUBEPLUS_NAMESPACE="$2"
                shift
            fi
            ;;
        *)
            usage
            ;;
    esac
    shift
done

# Function to configure provider kubeconfig for KubePlus
provider_kubeconfig() {
    echo "Downloading and setting up provider-kubeconfig.py"
    kubectl create namespace $KUBEPLUS_NAMESPACE 2>/dev/null || echo "Namespace $KUBEPLUS_NAMESPACE already exists"

    if ! $KUBEPLUS_CI; then
        wget -q https://raw.githubusercontent.com/cloud-ark/kubeplus/master/requirements.txt || { echo "Failed to download requirements.txt"; exit 1; }
        wget -q https://raw.githubusercontent.com/cloud-ark/kubeplus/master/provider-kubeconfig.py || { echo "Failed to download provider-kubeconfig.py"; exit 1; }
    fi
    python3 -m venv venv
    source venv/bin/activate
    pip3 install -r requirements.txt
    apiserver=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
    python3 provider-kubeconfig.py -s $apiserver create $KUBEPLUS_NAMESPACE
    deactivate
}

if $INSTALL_KUBEPLUS_PLUGIN; then
    echo "Installing KubePlus Plugin"
    if ! $KUBEPLUS_CI; then
        wget -q https://github.com/cloud-ark/kubeplus/raw/master/kubeplus-kubectl-plugins.tar.gz || { echo "Failed to download kubeplus-kubectl-plugins.tar.gz"; exit 1; }
        tar -zxvf kubeplus-kubectl-plugins.tar.gz
    fi
    export KUBEPLUS_HOME=$(pwd)
    export PATH=$KUBEPLUS_HOME/plugins:$PATH
    kubectl kubeplus commands
fi

# Install KubePlus if the flag is provided
if $INSTALL_KUBEPLUS; then
    provider_kubeconfig
    echo "Installing KubePlus in namespace: $KUBEPLUS_NAMESPACE"

    if ! $KUBEPLUS_CI; then
        helm install kubeplus https://github.com/cloud-ark/operatorcharts/raw/master/kubeplus-chart-4.0.0.tgz --kubeconfig=kubeplus-saas-provider.json -n $KUBEPLUS_NAMESPACE
    else
        helm install kubeplus ./deploy/kubeplus-chart \
        --kubeconfig=kubeplus-saas-provider.json \
        --set MUTATING_WEBHOOK=gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:latest \
        --set PLATFORM_OPERATOR=gcr.io/cloudark-kubeplus/platform-operator:latest \
        --set HELMER=gcr.io/cloudark-kubeplus/helm-pod:latest \
        --set CRD_REGISTRATION_HELPER=gcr.io/cloudark-kubeplus/kubeconfiggenerator:latest \
        -n $KUBEPLUS_NAMESPACE
    fi

    elapsed=0
    while ! kubectl get pods --namespace $KUBEPLUS_NAMESPACE | grep kubeplus | grep Running; do
        if [ $elapsed -ge $TIMEOUT ]; then
            echo "Timed out waiting for KubePlus to start."
            exit 1
        fi
        echo "Waiting for KubePlus to start.."
        sleep 1
        elapsed=$((elapsed + 1))
    done
    echo "KubePlus has started successfully."
fi

# Provide additional instructions or notes
if ! $INSTALL_KUBEPLUS && ! $INSTALL_KUBEPLUS_PLUGIN; then
    echo "No installation flags were provided. Please specify at least one installation flag: --kubeplus, or --kubeplus-plugin."
else
    echo "Installation process completed. Check the status of the deployments using 'kubectl get pods -n <namespace>'."
fi