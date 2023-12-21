#!/bin/bash

current_context=`kubectl config current-context`
kubectl config view | grep -B1 $current_context | grep server | awk '{print $2}'
