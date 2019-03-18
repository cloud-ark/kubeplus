#!/bin/bash

retry() {
    local -r -i max_attempts="$1"; shift
    local -r cmd="$@"
    local -i attempt_num=1

    until $cmd
    do
        if (( attempt_num == max_attempts ))
        then
            echo "Attempt $attempt_num failed and there are no more attempts left!"
            return 1
        else
            echo "Attempt $attempt_num failed! Trying again in $attempt_num seconds..."
            sleep $(( attempt_num++ ))
        fi
    done
}

dir="/kubeplus/scripts"
#If I am running this within scripts directory
if [[ $PWD == *$dir ]]; then
    cd ../
    export PROJECT_HOME=$PWD
    cd scripts
else #otherwise I assume I am being run from project root
    export PROJECT_HOME=$PWD
fi
kubectl cluster-info
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}';
function kube_manager(){
    until kubectl -n kube-system get pods -lcomponent=kube-addon-manager -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 3; kubectl get pods -n default; echo "waiting for kube-addon-manager to be available"; kubectl get pods --all-namespaces; done
    echo "kube-addon-manager successfully running."
}
function kube_dns(){
    until kubectl -n kube-system get pods -lk8s-app=kube-dns -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 3; kubectl get pods -n default; echo "waiting for kube-dns to be available"; done
    echo "kube-dns successfully running."
    echo ""
}
function helm_init(){
    helm init
    until kubectl -n kube-system get pods -l app="helm",name="tiller" -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 3;echo "waiting for tiller pod to be available"; done
    echo ""
}
function create_kubeplus(){
    echo "Creating Kubeplus pod ..."
    kubectl apply -f $PROJECT_HOME/deploy
    timeout=60
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapiserver="true" -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")

    while ! [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            exit 1
        fi
        result=$(kubectl -n default get pods -lapiserver="true" -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        echo $result
        sleep $sleep_time
        echo "Waiting for KubePlus deployment..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Kubeplus pod successfully running..."
    echo ""
}
function create_operator(){
    echo "Creating postgres-operator ..."
    kubectl create -f $PROJECT_HOME/examples/postgres/postgres-operator.yaml
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=postgres-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep kubeplus | awk '{print $1'}`
    while ! [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to create a postgres operator."
            echo ""
            echo "Operator deployer logs: "
            kubectl logs $pod -c operator-deployer
            echo ""
            echo "Operator manager logs: "
            kubectl logs $pod -c operator-manager
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=postgres-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        echo $result
        sleep $sleep_time
        echo "Waiting for postgres-operator..."
        kubectl get pods -n default
        echo ""
        count=$((count + 1))
    done
    echo "Successfully deployed postgres-operator!"
    echo ""
}
function create_postgres(){
    echo "Creating postgres instance ..."
    kubectl create -f $PROJECT_HOME/examples/postgres/postgres1.yaml
    timeout=600
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=postgres1  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep -v postgres-operator | grep postgres | awk '{print $1}'`
    while ! [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to create postgres."
            echo ""
            echo "Operator logs: "
            kubectl logs $pod
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=postgres1  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        echo $result
        sleep $sleep_time
        echo "Creating postgres ..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Successfully created postgres instance!"
    echo ""
}
function delete_postgres() {
    echo "Deleting postgres instance ..."
    kubectl delete -f $PROJECT_HOME/examples/postgres/postgres1.yaml
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=postgres1  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep -v postgres-operator | grep postgres | awk '{print $1}'`
    while [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to delete postgres."
            echo ""
            echo "Operator logs: "
            kubectl logs $pod
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=postgres1  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        sleep $sleep_time
        echo "Deleting postgres..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Successfully deleted postgres instance!"
    echo ""
}
function delete_operator() {
    echo "Deleting postgres operator ..."
    kubectl delete -f $PROJECT_HOME/examples/postgres/postgres-operator.yaml
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=postgres-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep kubeplus | awk '{print $1'}`

    while [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to delete postgres operator."
            echo ""
            echo "Operator deployer logs: "
            kubectl logs $pod -c operator-deployer
            echo ""
            echo "Operator manager logs: "
            kubectl logs $pod -c operator-manager
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=postgres-operator -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        sleep $sleep_time
        echo "Deleting postgres operator..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Successfully deleted postgres operator!"
    echo ""
}
function delete_kubeplus() {
    echo "Deleting Kubeplus pod ..."
    kubectl delete -f $PROJECT_HOME/deploy
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapiserver="true"  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    while [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            exit 1
        fi
        result=$(kubectl -n default get pods -lapiserver="true" -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        sleep $sleep_time
        echo "Deleting kubeplus..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Successfully deleted Kubeplus pod!"
    echo ""
}

declare -fxr create_kubeplus
declare -fxr create_operator
declare -fxr create_postgres
declare -fxr delete_postgres
declare -fxr delete_operator
declare -fxr delete_kubeplus


kube_manager
kube_dns
helm_init
retry 3 create_kubeplus
retry 3 create_operator
retry 3 create_postgres
retry 3 delete_postgres
retry 3 delete_operator
retry 3 delete_kubeplus
echo "TESTS PASSED! No non-zero error codes."
