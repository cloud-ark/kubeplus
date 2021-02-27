#!/bin/bash -x

namespace=$1
dependency=$2

while [ True ]
do
   status=`./root/kubectl get pods -n $namespace | grep $dependency | awk '{print $3}'`    
   if [[ $status != 'Running' ]]; then
    echo $status
    sleep 30
   else
     break
   fi 
done
echo "Done waiting.."
