#!/bin/bash

current_context=`kubectl config current-context`
current_context1=`echo $current_context | cut -d @ -f 2`
kubectl config view | grep -B1 $current_context1 | grep server | awk '{print $2}'
