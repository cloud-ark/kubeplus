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
    echo "Creating moodle-operator ..."
    helm install https://github.com/cloud-ark/operatorcharts/blob/master/moodle-operator-chart-0.3.0.tgz?raw=true
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=moodle-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep kubeplus | awk '{print $1'}`
    while ! [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to create moodle operator."
            echo ""
            echo "kube discovery logs: "
            kubectl logs $pod -c kube-discovery-apiserver
            echo ""
            echo "discovery helper logs: "
            kubectl logs $pod -c operator-discovery-helper
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=moodle-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        echo $result
        sleep $sleep_time
        echo "Waiting for moodle-operator..."
        kubectl get pods -n default
        echo ""
        count=$((count + 1))
    done
    echo "Successfully deployed moodle-operator!"
    echo ""
}
function test_explain() {
    echo "Testing explain endpoint ..."
    kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle"  | python -m json.tool
    resp=`curl -i "http://localhost:8080/apis/platform-as-code/v1/explain?kind=Moodle" | sed -e 1q | awk '{print $2}'`
    if [ $resp -eq 200 ]; then
        echo "Successfully called explain endpoint!"
        echo ""
    else
        echo "Unable to call explain endpoint!"
        echo ""
        exit 1
    fi

    kubectl get --raw "/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"  | python -m json.tool
    curl -i "http://localhost:8080/apis/platform-as-code/v1/explain?kind=Moodle.MoodleSpec"
    if [ $? -eq 0 ]; then
        echo "Successfully called explain endpoint!"
        echo ""
    else
        echo "Unable to call explain endpoint!"
        echo ""
        exit 1
    fi

}
function test_man() {
    echo "Testing man endpoint ..."
    kubectl get --raw "/apis/platform-as-code/v1/man?kind=Moodle"
    resp=`curl -i "http://localhost:8080/apis/platform-as-code/v1/man?kind=Moodle" | sed -e 1q | awk '{print $2}'`
    if [ $resp -eq 200 ]; then
        echo "Successfully called man endpoint!"
        echo ""
    else
        echo "Unable to call man endpoint!"
        echo ""
        exit 1
    fi
}
function delete_operator() {
    echo "Deleting moodle operator ..."
    operator=`helm list| grep moodle-operator-chart-0.3.0 | awk '{print $1'}`
    helm delete $operator --purge
    if [ $? -ne 0 ]; then
        echo "Unable to delete moodle operator."
        exit 1
    fi
    timeout=240
    sleep_time=3
    loops=$((timeout/sleep_time))
    count=0
    result=$(kubectl -n default get pods -lapp=moodle-operator  -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
    pod=`kubectl get pods | grep kubeplus | awk '{print $1'}`

    while [[ "$result" ]]; do
        if [[ "$count" -gt $loops ]]; then
            echo "Unable to delete moodle operator."
            echo ""
            echo "kube discovery logs: "
            kubectl logs $pod -c kube-discovery-apiserver
            echo ""
            echo "discovery helper logs: "
            kubectl logs $pod -c operator-discovery-helper
            exit 1
        fi
        result=$(kubectl -n default get pods -lapp=moodle-operator -o jsonpath="$JSONPATH" 2>&1 | grep "Ready=True")
        sleep $sleep_time
        echo "Deleting moodle operator..."
        kubectl get pods -n default
        count=$((count + 1))
    done
    echo "Successfully deleted moodle operator!"
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
declare -fxr test_explain
declare -fxr test_man
declare -fxr delete_operator
declare -fxr delete_kubeplus


kube_manager
kube_dns
helm_init
kubectl proxy --port=8080 &
retry 3 create_kubeplus
retry 3 create_operator
retry 3 test_explain
retry 3 test_man
retry 3 delete_operator
retry 3 delete_kubeplus

echo "TESTS PASSED! No non-zero error codes."
